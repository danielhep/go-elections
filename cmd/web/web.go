package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/danielhep/go-elections/internal/database"
	"github.com/danielhep/go-elections/internal/types"
	"github.com/gorilla/mux"
)

//go:embed images
var staticFiles embed.FS

func main() {
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		log.Fatal("PG_URL environment variable is not set")
	}

	// Connect to the database
	db, err := database.NewDB(pgURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := mux.NewRouter()

	// Main page route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var contests []types.Contest
		if err := db.Find(&contests).Error; err != nil {
			http.Error(w, "Error fetching contests", http.StatusInternalServerError)
			return
		}
		err = mainPage(contests).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Contest details route
	r.HandleFunc("/contest/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		contestID := vars["id"]

		var contest types.Contest
		if err := db.First(&contest, contestID).Error; err != nil {
			http.Error(w, "Contest not found", http.StatusNotFound)
			return
		}

		var candidates []types.BallotResponse
		if err := db.Where("contest_id = ?", contestID).
			Preload("VoteTallies").
			Preload("VoteTallies.Update").
			Find(&candidates).Error; err != nil {
			http.Error(w, "Error fetching vote tallies", http.StatusInternalServerError)
			return
		}
		// Sort candidates by their latest vote count
		sortCandidatesByLatestVotes(candidates)

		// Get the countyUpdates sorted
		var countyUpdates []types.Update
		var stateUpdate *types.Update
		for _, voteTally := range candidates[0].VoteTallies {
			if voteTally.Update.JurisdictionType == types.StateJurisdiction {
				stateUpdate = &voteTally.Update
			} else {
				countyUpdates = append(countyUpdates, voteTally.Update)
			}
		}
		sort.Slice(countyUpdates, func(a, b int) bool {
			return countyUpdates[b].Timestamp.Before(countyUpdates[a].Timestamp)
		})

		err = contestPage(contest, candidates, countyUpdates, stateUpdate).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	staticFS, err := fs.Sub(staticFiles, "images")
	if err != nil {
		log.Fatal(err)
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
