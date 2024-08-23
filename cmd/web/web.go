package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
		if err := db.Model(&types.Contest{}).Find(&contests).Error; err != nil {
			http.Error(w, "Error fetching contests", http.StatusInternalServerError)
			return
		}
		fmt.Printf("%#v\n", contests)
		err = mainPage(contests).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Contest details route
	r.HandleFunc("/contest/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		contestID := vars["id"]
		timestamp := r.URL.Query().Get("timestamp")

		var contest types.Contest
		if err := db.First(&contest, contestID).Error; err != nil {
			http.Error(w, "Contest not found", http.StatusNotFound)
			return
		}

		var update types.Update
		if timestamp == "" {
			// Get the most recent update
			if err := db.Order("timestamp desc").First(&update).Error; err != nil {
				http.Error(w, "Error fetching latest update", http.StatusInternalServerError)
				return
			}
		} else {
			// Get the update for the specified timestamp
			if err := db.Where("timestamp = ?", timestamp).First(&update).Error; err != nil {
				http.Error(w, "Update not found for the specified timestamp", http.StatusNotFound)
				return
			}
		}

		var voteTallies []types.VoteTally
		if err := db.Where("contest_id = ? AND update_id = ?", contestID, update.ID).
			Preload("Candidate").Find(&voteTallies).Error; err != nil {
			http.Error(w, "Error fetching vote tallies", http.StatusInternalServerError)
			return
		}

		// Fetch all available timestamps for the dropdown
		var updates []types.Update
		if err := db.Order("timestamp desc").Find(&updates).Error; err != nil {
			http.Error(w, "Error fetching timestamps", http.StatusInternalServerError)
			return
		}

		err = contestPage(contest, voteTallies, updates, update.Timestamp).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
