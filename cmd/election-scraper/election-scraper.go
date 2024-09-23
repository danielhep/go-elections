package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danielhep/go-elections/internal/csv"
	"github.com/danielhep/go-elections/internal/database"
	"github.com/danielhep/go-elections/internal/types"
)

func initalLoad(db *database.DB) error {
	stateURL := os.Getenv("STATE_DATA")
	countyURL := os.Getenv("COUNTY_DATA")
	election, err := db.GetElection()
	if err != nil {
		return fmt.Errorf("error getting election: %v", err)
	}
	data, hash, err := csv.ParseFromURL(stateURL, types.StateJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", types.StateJurisdiction, err)
	}
	if err := db.LoadBallotResponses(data, *election); err != nil {
		return err
	}
	if err := db.CheckAndProcessUpdate(data, hash, types.StateJurisdiction, *election); err != nil {
		return err
	}

	data, hash, err = csv.ParseFromURL(countyURL, types.CountyJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", types.CountyJurisdiction, err)
	}
	if err := db.LoadBallotResponses(data, *election); err != nil {
		return err
	}
	if err := db.CheckAndProcessUpdate(data, hash, types.CountyJurisdiction, *election); err != nil {
		return err
	}
	return nil
}

// Function to check for updates
func checkForUpdates(db *database.DB) error {
	stateURL := os.Getenv("STATE_DATA")
	countyURL := os.Getenv("COUNTY_DATA")
	election, err := db.GetElection()
	if err != nil {
		return fmt.Errorf("error getting election: %v", err)
	}

	data, hash, err := csv.ParseFromURL(stateURL, types.StateJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", types.StateJurisdiction, err)
	}
	if err := db.CheckAndProcessUpdate(data, hash, types.StateJurisdiction, *election); err != nil {
		return err
	}

	data, hash, err = csv.ParseFromURL(countyURL, types.CountyJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", types.CountyJurisdiction, err)
	}
	if err := db.CheckAndProcessUpdate(data, hash, types.CountyJurisdiction, *election); err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println("Election data")
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		log.Fatal("PG_URL environment variable is not set")
	}

	// Connect to the database
	db, err := database.NewDB(pgURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate the schema
	err = db.MigrateSchema()
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	// Set up a ticker to periodically check for updates
	updateInterval := time.Second
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Run the first check immediately
	if err := initalLoad(db); err != nil {
		log.Printf("Error checking for updates: %v", err)
	}

	// Set up a channel to handle graceful shutdown
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := checkForUpdates(db); err != nil {
					log.Printf("Error checking for updates: %v", err)
				}
			case <-done:
				return
			}
		}
	}()

	// Keep the main goroutine running
	fmt.Println("Update checker is running. Press Ctrl+C to stop.")
	select {}
}
