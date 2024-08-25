package main

import (
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/danielhep/go-elections/internal/database"
	"github.com/danielhep/go-elections/internal/types"
	"github.com/gorilla/mux"
)

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

		var candidates []types.Candidate
		if err := db.Where("contest_id = ?", contestID).
			Preload("VoteTallies").
			Preload("VoteTallies.Update").
			Find(&candidates).Error; err != nil {
			http.Error(w, "Error fetching vote tallies", http.StatusInternalServerError)
			return
		}
		// Sort candidates by their latest vote count
		sortCandidatesByLatestVotes(candidates)

		// Get the updates sorted
		var updates []types.Update
		for _, voteTally := range candidates[0].VoteTallies {
			updates = append(updates, voteTally.Update)
		}
		sort.Slice(updates, func(a, b int) bool {
			// put state at the top
			return updates[b].Timestamp.Before(updates[a].Timestamp)
		})

		err = contestPage(contest, candidates, updates).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
