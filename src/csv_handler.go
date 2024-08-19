package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-pg/pg/v10"
)

// Function to scrape and parse CSV data
func scrapeAndParse(url string) ([][]string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	hash := calculateHash(body)

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, "", err
	}
	return records, hash, nil
}

// Function to calculate hash of CSV data
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Function to check and process updates for a specific jurisdiction
func checkAndProcessUpdate(db *pg.DB, url string, jurisdictionType JurisdictionType) error {
	data, hash, err := scrapeAndParse(url)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", jurisdictionType, err)
	}

	var lastUpdate Update
	err = db.Model(&lastUpdate).
		Where("type = ?", jurisdictionType).
		Order("id DESC").
		Limit(1).
		Select()

	if err == pg.ErrNoRows {
		log.Printf("First %s update detected", jurisdictionType)
		if err := updateDatabase(db, data, jurisdictionType, hash); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if err != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, err)
	} else if lastUpdate.Hash != hash {
		log.Printf("%s data change detected", jurisdictionType)
		if err := updateDatabase(db, data, jurisdictionType, hash); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else {
		log.Printf("No change in %s data", jurisdictionType)
	}

	return nil
}

// Function to update database
func updateDatabase(db *pg.DB, data [][]string, jurisdictionType JurisdictionType, hash string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Ensure rollback if panic occurs
	defer func() {
		if r := recover(); r != nil {
			err := tx.Rollback()
			if err != nil {
				log.Printf("failed to rollback transaction: %v", err)
			}
			panic(r) // re-throw panic after Rollback
		}
	}()

	// Create a new Update record
	update := &Update{
		Timestamp:        time.Now().Format(time.RFC3339),
		Hash:             hash,
		JurisdictionType: jurisdictionType,
	}
	_, err = tx.Model(update).Insert()
	if err != nil {
		return err
	}

	_ = data

	// TODO: Process the CSV data and update the database tables
	// This will involve parsing the CSV data and updating the Contest, Candidate, and VoteTally tables

	// Commit the transaction
	return tx.Commit()
}
