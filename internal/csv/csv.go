package csv

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/danielhep/go-elections/internal/types"
	"github.com/gocarina/gocsv"
	"gorm.io/gorm"
)

func Parse(reader io.ReadCloser, jurisdictionType types.JurisdictionType) ([]types.GenericVoteRecord, string, error) {
	// Create a TeeReader to read the body and calculate hash simultaneously
	hashReader := sha256.New()
	teeReader := io.TeeReader(reader, hashReader)

	var records []types.GenericVoteRecord
	switch jurisdictionType {
	case types.StateJurisdiction:
		var stateRecords []*types.StateCSVRecord
		if err := gocsv.Unmarshal(teeReader, &stateRecords); err != nil {
			return nil, "", err
		}
		for _, record := range stateRecords {
			records = append(records, record.ToGeneric())
		}
	case types.CountyJurisdiction:
		var countyRecords []*types.CountyCSVRecord
		if err := gocsv.Unmarshal(teeReader, &countyRecords); err != nil {
			return nil, "", err
		}
		for _, record := range countyRecords {
			records = append(records, record.ToGeneric())
		}
	default:
		return nil, "", fmt.Errorf("unknown jurisdiction type: %s", jurisdictionType)
	}

	// Calculate hash
	hash := hex.EncodeToString(hashReader.Sum(nil))

	return records, hash, nil
}

// Function to scrape and parse CSV data
func ParseFromURL(url string, jurisdictionType types.JurisdictionType) ([]types.GenericVoteRecord, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	return Parse(resp.Body, jurisdictionType)
}

// Function to check and process updates for a specific jurisdiction
func CheckAndProcessUpdate(db *gorm.DB, data []types.GenericVoteRecord, hash string, jurisdictionType types.JurisdictionType) error {
	var update types.Update
	result := db.Where("hash = ?", hash).First(&update)
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("New %s update detected", jurisdictionType)
		if err := UpdateVoteTallies(db, data, jurisdictionType, hash, time.Now()); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, result.Error)
	} else {
		log.Printf("No change in %s data", jurisdictionType)
	}

	return nil
}

func LoadCandidates(db *gorm.DB, data []types.GenericVoteRecord) error {
	// Process the data based on jurisdiction type
	var contests []types.Contest
	var err error

	contests, err = processContests(data)

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

	fmt.Printf("Loading %v contests.\n", len(contests))

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
				return fmt.Errorf("error creating candidate: %v \n under contest %+v", err, contest)
			}
		}
	}

	fmt.Printf("Total candidates: %v\n", totalCandidates)

	// Commit the transaction
	return tx.Commit().Error
}

// Function to update database
func UpdateVoteTallies(db *gorm.DB, data []types.GenericVoteRecord, jurisdictionType types.JurisdictionType, hash string, timestamp time.Time) error {
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
	update := &types.Update{
		Timestamp:        timestamp.Format(time.RFC3339),
		Hash:             hash,
		JurisdictionType: jurisdictionType,
	}
	if err := tx.Create(update).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Preload existing contests and candidates
	var contests []types.Contest
	var candidates []types.Candidate
	if err := tx.Find(&contests).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Find(&candidates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create maps for quick lookups
	contestMap := make(map[string]uint)
	for _, c := range contests {
		key := getContestKey(c.Name, c.District)
		contestMap[key] = c.ID
	}

	candidateMap := make(map[string]uint)
	for _, c := range candidates {
		key := fmt.Sprintf("%d-%s", c.ContestID, c.Name)
		candidateMap[key] = c.ID
	}
	// Process vote tallies
	var voteTallies []types.VoteTally
	for _, record := range data {
		contestKey := getContestKey(record.BallotTitle, record.DistrictName)
		contestID, contestExists := contestMap[contestKey]
		if !contestExists {
			tx.Rollback()
			return fmt.Errorf("contest not found: %s", contestKey)
		}
		candidateKey := fmt.Sprintf("%d-%s", contestID, record.BallotResponse)
		candidateID, candidateExists := candidateMap[candidateKey]
		if !candidateExists {
			tx.Rollback()
			return fmt.Errorf("candidate not found: %s", candidateKey)
		}
		// Create vote tally
		voteTally := types.VoteTally{
			CandidateID: candidateID,
			UpdateID:    update.ID,
			Votes:       record.Votes,
			ContestID:   contestID,
		}
		voteTallies = append(voteTallies, voteTally)
	}
	// Insert vote tallies in batches
	if len(voteTallies) > 0 {
		fmt.Printf("Loading %v vote tallies for %v\n", len(voteTallies), jurisdictionType)
		if err := tx.CreateInBatches(voteTallies, 100).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating vote tallies: %v", err)
		}
	}

	return tx.Commit().Error
}

// Function to process state-level data
func processContests(records []types.GenericVoteRecord) ([]types.Contest, error) {
	contestMap := make(map[string]*types.Contest)

	for _, record := range records {
		// Create or get Contest
		contestKey := fmt.Sprintf("%s-%s", record.BallotTitle, record.DistrictName)
		contest, exists := contestMap[contestKey]
		if !exists {
			contest = &types.Contest{
				Name:       record.BallotTitle,
				District:   record.DistrictName,
				Candidates: []types.Candidate{},
			}
			contestMap[contestKey] = contest
		}

		// Create Candidate and add to Contest
		candidate := types.Candidate{
			Name:  record.BallotResponse,
			Party: &record.PartyPreference,
		}
		contest.Candidates = append(contest.Candidates, candidate)
	}
	// Convert map to slice
	contests := make([]types.Contest, 0, len(contestMap))
	for _, contest := range contestMap {
		contests = append(contests, *contest)
	}

	return contests, nil
}

func getContestKey(ballotTitle string, districtName string) string {
	return fmt.Sprintf("%s-%s", ballotTitle, districtName)
}
