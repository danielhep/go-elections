package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type JurisdictionType string

const (
	StateJurisdiction  JurisdictionType = "State"
	CountyJurisdiction JurisdictionType = "County"
)

// Structs to represent the data
type Contest struct {
	ID               int              `pg:"id,pk"`
	Name             string           `pg:"name,notnull"`
	JurisdictionType JurisdictionType `pg:"type,notnull"` // "State" or "County"
}

type Candidate struct {
	ID        int      `pg:"id,pk"`
	Name      string   `pg:"name,notnull"`
	Party     string   `pg:"party"`
	ContestID int      `pg:"contest_id,notnull"`
	Contest   *Contest `pg:"rel:has-one"`
}

type Update struct {
	ID               int              `pg:"id,pk"`
	Timestamp        string           `pg:"timestamp,notnull"`
	Hash             string           `pg:"hash,notnull"`
	JurisdictionType JurisdictionType `pg:"type,notnull"` // "State" or "County"
}

type VoteTally struct {
	ID          int        `pg:"id,pk"`
	CandidateID int        `pg:"candidate_id,notnull"`
	Candidate   *Candidate `pg:"rel:has-one"`
	UpdateID    int        `pg:"update_id,notnull"`
	Update      *Update    `pg:"rel:has-one"`
	Votes       int        `pg:"votes,notnull"`
}

// Function to check for updates
func checkForUpdates(db *pg.DB) error {
	stateURL := os.Getenv("STATE_DATA")
	countyURL := os.Getenv("COUNTY_DATA")

	if err := checkAndProcessUpdate(db, stateURL, StateJurisdiction); err != nil {
		return err
	}

	if err := checkAndProcessUpdate(db, countyURL, CountyJurisdiction); err != nil {
		return err
	}

	return nil
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*Contest)(nil),
		(*Candidate)(nil),
		(*Update)(nil),
		(*VoteTally)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
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

	// Parse the connection string and connect to the database
	opt, err := pg.ParseURL(pgURL)
	if err != nil {
		log.Fatalf("Failed to parse PG_URL: %v", err)
	}

	db := pg.Connect(opt)
	defer db.Close()

	// Check the connection
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	// Create the schema
	if err := createSchema(db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Schema created successfully")

	// Set up a ticker to periodically check for updates
	updateInterval := 5 * time.Minute // Check every 5 minutes
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Run the first check immediately
	if err := checkForUpdates(db); err != nil {
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
