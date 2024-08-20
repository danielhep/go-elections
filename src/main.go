package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JurisdictionType string

const (
	StateJurisdiction  JurisdictionType = "State"
	CountyJurisdiction JurisdictionType = "County"
)

// Structs to represent the data
type Contest struct {
	gorm.Model
	Name             string
	District         string
	JurisdictionType JurisdictionType
	Candidates       []Candidate
}

type Candidate struct {
	gorm.Model
	Name      string
	Party     string
	ContestID uint
	Contest   Contest
}

type Update struct {
	gorm.Model
	Timestamp        string
	Hash             string
	JurisdictionType JurisdictionType
	VoteTallies      []VoteTally
}

type VoteTally struct {
	gorm.Model
	CandidateID uint
	Candidate   Candidate
	UpdateID    uint
	Update      Update
	Votes       int
}

func initalLoad(db *gorm.DB) error {
	stateURL := os.Getenv("STATE_DATA")
	countyURL := os.Getenv("COUNTY_DATA")
	data, hash, err := scrapeAndParse(stateURL, StateJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", StateJurisdiction, err)
	}
	if err := loadCandidates(db, data, StateJurisdiction); err != nil {
		return err
	}
	if err := checkAndProcessUpdate(db, data, hash, StateJurisdiction); err != nil {
		return err
	}

	data, hash, err = scrapeAndParse(countyURL, CountyJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", CountyJurisdiction, err)
	}
	if err := loadCandidates(db, data, CountyJurisdiction); err != nil {
		return err
	}
	if err := checkAndProcessUpdate(db, data, hash, CountyJurisdiction); err != nil {
		return err
	}
	return nil
}

// Function to check for updates
func checkForUpdates(db *gorm.DB) error {
	stateURL := os.Getenv("STATE_DATA")
	countyURL := os.Getenv("COUNTY_DATA")

	data, hash, err := scrapeAndParse(stateURL, StateJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", StateJurisdiction, err)
	}
	if err := checkAndProcessUpdate(db, data, hash, StateJurisdiction); err != nil {
		return err
	}

	data, hash, err = scrapeAndParse(countyURL, CountyJurisdiction)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", CountyJurisdiction, err)
	}
	if err := checkAndProcessUpdate(db, data, hash, CountyJurisdiction); err != nil {
		return err
	}

	return nil
}

func main() {
	// Get the connection string from the environment variable
	fmt.Println("Election data")
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		log.Fatal("PG_URL environment variable is not set")
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate the schema
	err = db.AutoMigrate(&Contest{}, &Candidate{}, &Update{}, &VoteTally{})
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
	fmt.Println("Schema migrated successfully")

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
