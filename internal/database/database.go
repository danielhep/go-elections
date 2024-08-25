package database

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/danielhep/go-elections/internal/csv"
	"github.com/danielhep/go-elections/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(pgURL string) (*DB, error) {
	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) MigrateSchema() error {
	err := db.AutoMigrate(&types.Contest{}, &types.Candidate{}, &types.Update{}, &types.VoteTally{})
	if err != nil {
		return fmt.Errorf("failed to migrate database schema: %v", err)
	}
	log.Println("Schema migrated successfully")
	return nil
}

func (db *DB) LoadCandidates(data []types.GenericVoteRecord) error {
	// Process the data based on jurisdiction type
	var contests []types.Contest
	var err error

	contests, err = csv.ProcessContests(data)

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

// Creates an update entry in the database and then creates a VoteTally entry for
// every entry in the GenericVoteRecord.
func (db *DB) UpdateVoteTallies(data []types.GenericVoteRecord, hash string, timestamp time.Time) error {
	jType := data[0].JurisdictionType
	if slices.ContainsFunc(data, func(entry types.GenericVoteRecord) bool { return entry.JurisdictionType != data[0].JurisdictionType }) {
		return fmt.Errorf("Error, found inconsistent jurisdiction types while updating vote tallies.")
	}
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
		JurisdictionType: jType,
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
		fmt.Printf("Loading %v vote tallies for %v\n", len(voteTallies), jType)
		if err := tx.CreateInBatches(voteTallies, 100).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating vote tallies: %v", err)
		}
	}

	return tx.Commit().Error
}

// Checks the hash and publishes a new update if the has doesn't exist yet
func (db *DB) CheckAndProcessUpdate(data []types.GenericVoteRecord, hash string, jurisdictionType types.JurisdictionType) error {
	var update types.Update
	result := db.Where("hash = ?", hash).First(&update)
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("New %s update detected", jurisdictionType)
		if err := db.UpdateVoteTallies(data, hash, time.Now()); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, result.Error)
	} else {
		log.Printf("No change in %s data", jurisdictionType)
	}

	return nil
}

func getContestKey(ballotTitle string, districtName string) string {
	return fmt.Sprintf("%s-%s", ballotTitle, districtName)
}
