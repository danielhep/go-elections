package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// StateCSVRecord represents the structure of each row in the state CSV file
type StateCSVRecord struct {
	Race                   string  `csv:"Race"`
	Candidate              string  `csv:"Candidate"`
	Party                  string  `csv:"Party"`
	Votes                  int     `csv:"Votes"`
	PercentageOfTotalVotes float64 `csv:"PercentageOfTotalVotes"`
	JurisdictionName       string  `csv:"JurisdictionName"`
}

// CountyCSVRecord represents the structure of each row in the county CSV file
type CountyCSVRecord struct {
	GEMSContestID               string  `csv:"GEMS Contest ID"`
	ContestSortSeq              int     `csv:"Contest Sort Seq"`
	DistrictType                string  `csv:"District Type"`
	DistrictTypeSubheading      string  `csv:"District Type Subheading"`
	DistrictName                string  `csv:"District Name"`
	BallotTitle                 string  `csv:"Ballot Title"`
	BallotsCountedForDistrict   int     `csv:"Ballots Counted for District"`
	RegisteredVotersForDistrict int     `csv:"Registered Voters for District"`
	PercentTurnoutForDistrict   float64 `csv:"Percent Turnout for District"`
	CandidateSortSeq            int     `csv:"Candidate Sort Seq"`
	BallotResponse              string  `csv:"Ballot Response"`
	PartyPreference             string  `csv:"Party Preference"`
	Votes                       int     `csv:"Votes"`
	PercentOfVotes              float64 `csv:"Percent of Votes"`
}

// Function to scrape and parse CSV data
func scrapeAndParse(url string, jurisdictionType JurisdictionType) (interface{}, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Create a TeeReader to read the body and calculate hash simultaneously
	hashReader := sha256.New()
	teeReader := io.TeeReader(resp.Body, hashReader)

	var records interface{}
	switch jurisdictionType {
	case StateJurisdiction:
		var stateRecords []*StateCSVRecord
		if err := gocsv.Unmarshal(teeReader, &stateRecords); err != nil {
			return nil, "", err
		}
		records = stateRecords
	case CountyJurisdiction:
		var countyRecords []*CountyCSVRecord
		if err := gocsv.Unmarshal(teeReader, &countyRecords); err != nil {
			return nil, "", err
		}
		records = countyRecords
	default:
		return nil, "", fmt.Errorf("unknown jurisdiction type: %s", jurisdictionType)
	}

	// Calculate hash
	hash := hex.EncodeToString(hashReader.Sum(nil))

	return records, hash, nil
}

// Function to check and process updates for a specific jurisdiction
func checkAndProcessUpdate(db *gorm.DB, data interface{}, hash string, jurisdictionType JurisdictionType) error {
	var lastUpdate Update
	result := db.Where("jurisdiction_type = ?", jurisdictionType).Order("id DESC").First(&lastUpdate)
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("First %s update detected", jurisdictionType)
		if err := updateVoteTallies(db, data, jurisdictionType, hash); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, result.Error)
	} else if lastUpdate.Hash != hash {
		log.Printf("%s data change detected", jurisdictionType)
		if err := updateVoteTallies(db, data, jurisdictionType, hash); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else {
		log.Printf("No change in %s data", jurisdictionType)
	}

	return nil
}

func loadCandidates(db *gorm.DB, data interface{}, jurisdictionType JurisdictionType) error {
	// Process the data based on jurisdiction type
	var contests []Contest
	var err error

	switch jurisdictionType {
	case StateJurisdiction:
		contests, err = processStateData(data.([]*StateCSVRecord))
	case CountyJurisdiction:
		contests, err = processCountyData(data.([]*CountyCSVRecord))
	default:
		err = fmt.Errorf("unknown jurisdiction type: %s", jurisdictionType)
	}

	tx := db.Begin()
	// Ensure rollback if panic occurs
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after Rollback
		}
	}()
	if err != nil {
		tx.Rollback()
		return err
	}

	fmt.Printf("Loading %v contests for %v.\n", len(contests), jurisdictionType)

	totalCandidates := 0
	// Insert contests and candidates into the database
	for _, contest := range contests {
		if err := tx.FirstOrCreate(&contest, contest).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating contest: %v", err)
		}
		totalCandidates += len(contest.Candidates)
		for _, candidate := range contest.Candidates {
			if err := tx.FirstOrCreate(&candidate, candidate).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating candidate: %v", err)
			}
		}
	}

	fmt.Printf("Total candidates: %v\n", totalCandidates)

	// Commit the transaction
	return tx.Commit().Error
}

// Function to update database
func updateVoteTallies(db *gorm.DB, data interface{}, jurisdictionType JurisdictionType, hash string) error {
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

	_ = data

	return tx.Commit().Error
}

// Function to process state-level data
func processStateData(records []*StateCSVRecord) ([]Contest, error) {
	contestMap := make(map[string]*Contest)

	for _, record := range records {
		contestName, district := extractContestInfo(record.Race)
		candidateName := normalizeString(record.Candidate)
		party := extractParty(record.Party)

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

// Function to process county-level data
func processCountyData(records []*CountyCSVRecord) ([]Contest, error) {
	contestMap := make(map[string]*Contest)

	for _, row := range records {
		contestName := normalizeString(row.BallotTitle)
		district := normalizeString(row.DistrictName)
		candidateName := normalizeString(row.BallotResponse)
		party := extractParty(row.PartyPreference)

		// Create or get Contest
		contestKey := fmt.Sprintf("%s-%s", contestName, district)
		contest, exists := contestMap[contestKey]
		if !exists {
			contest = &Contest{
				Name:             contestName,
				District:         district,
				JurisdictionType: CountyJurisdiction,
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
