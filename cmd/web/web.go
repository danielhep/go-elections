package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/danielhep/go-elections/internal"
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
	db, err := internal.NewDB(pgURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := mux.NewRouter()

	// Root page route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var elections []internal.Election
		if err := db.Order("election_date DESC").Find(&elections).Error; err != nil {
			http.Error(w, "Error fetching elections", http.StatusInternalServerError)
			return
		}
		err = rootPage(elections).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Election page route
	r.HandleFunc("/{electionID}/", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		electionID := vars["electionID"]
		var contests []internal.Contest
		if err := db.Where("election_id = ?", electionID).Preload("Election").Find(&contests).Error; err != nil {
			http.Error(w, "Error fetching contests", http.StatusInternalServerError)
			return
		}
		var election internal.Election
		if err := db.Where("id = ?", electionID).First(&election).Error; err != nil {
			http.Error(w, "Election not found", http.StatusNotFound)
			return
		}
		err = electionPage(election, contests).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Contest details route
	r.HandleFunc("/{electionID}/contest/{contestKey}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		contestKey := vars["contestKey"]
		electionID := vars["electionID"]

		contest := internal.Contest{
			ElectionID: electionID,
			ContestKey: contestKey,
		}
		if err := db.Preload("Election").First(&contest, contest).Error; err != nil {
			http.Error(w, "Contest not found", http.StatusNotFound)
			return
		}

		var candidates []internal.BallotResponse
		if err := db.Where("contest_id = ?", contest.ID).
			Preload("VoteTallies").
			Preload("VoteTallies.Update").
			Find(&candidates).Error; err != nil {
			http.Error(w, "Error fetching vote tallies", http.StatusInternalServerError)
			return
		}
		// Sort candidates by their latest vote count
		sortCandidatesByLatestVotes(candidates)

		// Get the countyUpdates sorted
		var countyUpdates []internal.Update
		var stateUpdate *internal.Update
		for _, voteTally := range candidates[0].VoteTallies {
			if voteTally.Update.JurisdictionType == internal.StateJurisdiction {
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
