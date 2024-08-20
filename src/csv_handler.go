package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// Function to scrape and parse CSV data
func scrapeAndParse(url string) ([][]string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Create a TeeReader to read the body and calculate hash simultaneously
	hashReader := sha256.New()
	teeReader := io.TeeReader(resp.Body, hashReader)

	// Create a buffered reader for CSV parsing
	bufReader := bufio.NewReader(teeReader)

	// Parse CSV
	csvReader := csv.NewReader(bufReader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, "", err
	}

	// Calculate hash
	hash := hex.EncodeToString(hashReader.Sum(nil))

	return records, hash, nil
}

// Function to check and process updates for a specific jurisdiction
func checkAndProcessUpdate(db *gorm.DB, url string, jurisdictionType JurisdictionType) error {
	data, hash, err := scrapeAndParse(url)
	if err != nil {
		return fmt.Errorf("error scraping %s data: %v", jurisdictionType, err)
	}

	var lastUpdate Update
	result := db.Where("jurisdiction_type = ?", jurisdictionType).Order("id DESC").First(&lastUpdate)
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("First %s update detected", jurisdictionType)
		if err := updateDatabase(db, data, jurisdictionType, hash); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, result.Error)
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
func updateDatabase(db *gorm.DB, data [][]string, jurisdictionType JurisdictionType, hash string) error {
	// Start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback if panic occurs
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after Rollback
		}
	}()

	// Create a new Update record
	update := &Update{
		Timestamp:        time.Now().Format(time.RFC3339),
		Hash:             hash,
		JurisdictionType: jurisdictionType,
	}
	if err := tx.Create(update).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Process the data based on jurisdiction type
	var contests []Contest
	var err error

	switch jurisdictionType {
	case StateJurisdiction:
		contests, err = processStateData(data)
	case CountyJurisdiction:
		// TODO: Implement county data processing
		err = fmt.Errorf("county data processing not implemented yet")
	default:
		err = fmt.Errorf("unknown jurisdiction type: %s", jurisdictionType)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert contests and candidates into the database
	for _, contest := range contests {
		if err := tx.FirstOrCreate(&contest, Contest{Name: contest.Name, District: contest.District, JurisdictionType: contest.JurisdictionType}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating contest: %v", err)
		}
		for _, candidate := range contest.Candidates {
			if err := tx.FirstOrCreate(&candidate, Candidate{Name: candidate.Name, Party: candidate.Party, ContestID: candidate.ContestID}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating candidate: %v", err)
			}
		}
	}

	// Commit the transaction
	return tx.Commit().Error
}

// Function to process state-level data
func processStateData(data [][]string) ([]Contest, error) {
	contestMap := make(map[string]*Contest)

	for i, row := range data {
		if i == 0 { // Skip header row
			continue
		}

		if len(row) < 5 {
			return nil, fmt.Errorf("row %d has insufficient columns", i)
		}

		contestName, district := extractContestInfo(row[0])
		candidateName := normalizeString(row[1])
		party := extractParty(row[2])

		// Create or get Contest
		contestKey := fmt.Sprintf("%s-%s", contestName, district)
		contest, exists := contestMap[contestKey]
		if !exists {
			contest = &Contest{
				Name:             contestName,
				District:         district,
				JurisdictionType: StateJurisdiction,
				Candidates:       []Candidate{},
			}
			contestMap[contestKey] = contest
		}

		// Create Candidate and add to Contest
		candidate := Candidate{
			Name:  candidateName,
			Party: party,
		}
		contest.Candidates = append(contest.Candidates, candidate)
	}
	// Convert map to slice
	contests := make([]Contest, 0, len(contestMap))
	for _, contest := range contestMap {
		contests = append(contests, *contest)
	}

	return contests, nil
}

// Helper function to extract contest name and district
func extractContestInfo(raceField string) (contestName string, district string) {
	parts := strings.SplitN(raceField, " - ", 2)
	if len(parts) == 2 {
		district = normalizeString(parts[0])
		contestName = normalizeString(parts[1])
	} else {
		contestName = normalizeString(raceField)
	}
	return
}

// Helper function to normalize string capitalization
func normalizeString(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, word := range words {
		if word != "of" && word != "the" && word != "and" && word != "in" && word != "for" {
			words[i] = cases.Title(language.AmericanEnglish).String(word)
		}
	}
	return strings.Join(words, " ")
}

// Helper function to extract party from the party field
func extractParty(partyField string) string {
	partyField = strings.TrimSpace(partyField)
	if strings.HasPrefix(partyField, "(") && strings.HasSuffix(partyField, ")") {
		partyField = partyField[1 : len(partyField)-1]
	}
	return normalizeString(strings.TrimPrefix(partyField, "Prefers "))
}
