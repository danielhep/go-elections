package internal

import (
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func NewDB(pgURL string) (*DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			IgnoreRecordNotFoundError: true, // Ignore ErrRecordNotFound error for logger
		},
	)

	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) MigrateSchema() error {
	err := db.AutoMigrate(&Contest{}, &BallotResponse{}, &Update{}, &VoteTally{})
	if err != nil {
		return fmt.Errorf("failed to migrate database schema: %v", err)
	}
	log.Println("Schema migrated successfully")
	return nil
}

func (db *DB) LoadBallotResponses(data []GenericVoteRecord, election Election) error {
	// Process the data based on jurisdiction type
	var contests []Contest
	var err error

	contests, err = ProcessContests(data, election)

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
		totalCandidates += len(contest.BallotResponses)
		for _, candidate := range contest.BallotResponses {
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
func (db *DB) UpdateVoteTallies(data []GenericVoteRecord, hash string, timestamp time.Time, election Election) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to process")
	}
	jType := data[0].JurisdictionType
	if slices.ContainsFunc(data, func(entry GenericVoteRecord) bool { return entry.JurisdictionType != data[0].JurisdictionType }) {
		return fmt.Errorf("error, found inconsistent jurisdiction types while updating vote tallies")
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
	update := &Update{
		Timestamp:        timestamp,
		Hash:             hash,
		JurisdictionType: jType,
		ElectionID:       election.ID,
	}
	if err := tx.Create(update).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Preload existing contests and candidates
	var contests []Contest
	var candidates []BallotResponse
	if err := tx.Find(&contests).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Find(&candidates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create maps for quick lookups
	contestMap := make(map[string]Contest)
	for _, c := range contests {
		key := getContestKey(c.BallotTitle, c.District)
		contestMap[key] = c
	}

	candidateMap := make(map[string]uint)
	for _, c := range candidates {
		key := getCandidateKey(c.ContestID, c.Name)
		candidateMap[key] = c.ID
	}
	// Process vote tallies
	var voteTallies []VoteTally
	for _, record := range data {
		contestKey := getContestKey(record.BallotTitle, record.DistrictName)
		contest, contestExists := contestMap[contestKey]
		if !contestExists {
			tx.Rollback()
			return fmt.Errorf("contest not found: %s", contestKey)
		}
		candidateKey := getCandidateKey(contest.ID, record.BallotResponse)
		ballotResponseID, candidateExists := candidateMap[candidateKey]
		if !candidateExists {
			tx.Rollback()
			return fmt.Errorf("candidate not found: %s", candidateKey)
		}
		// Create vote tally
		voteTally := VoteTally{
			BallotResponseID: ballotResponseID,
			UpdateID:         update.ID,
			Votes:            record.Votes,
			VotePercentage:   record.VotePercentage,
			ContestID:        contest.ID,
		}

		if !slices.Contains(contest.Jurisdictions, string(jType)) {
			contest.Jurisdictions = append(contest.Jurisdictions, string(jType))
			db.Model(&contest).Update("Jurisdictions", contest.Jurisdictions)
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

func (db *DB) UpdateHashExists(hash string) (bool, Update) {
	var update Update
	result := db.Where("hash = ?", hash).First(&update)
	return result.Error == nil, update
}

func (db *DB) DeleteUpdate(update Update) {
	db.Delete(&update)
}

// Checks the hash and publishes a new update if the has doesn't exist yet
func (db *DB) CheckAndProcessUpdate(data []GenericVoteRecord, hash string, jurisdictionType JurisdictionType, election Election) error {
	var update Update
	// Check to see if the update already exists
	result := db.Where("hash = ?", hash).First(&update)
	if result.Error == gorm.ErrRecordNotFound {
		log.Printf("New %s update detected", jurisdictionType)
		if err := db.UpdateVoteTallies(data, hash, time.Now(), election); err != nil {
			return fmt.Errorf("error updating %s data: %v", jurisdictionType, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("error querying %s update: %v", jurisdictionType, result.Error)
	} else {
		log.Printf("No change in %s data", jurisdictionType)
	}

	return nil
}

func (db *DB) GetElection() (*Election, error) {
	electionName := os.Getenv("ELECTION_NAME")
	electionDate, err := time.Parse("2006-01-02", os.Getenv("ELECTION_DATE"))
	if err != nil {
		return nil, fmt.Errorf("error parsing ELECTION_DATE %s: %v", os.Getenv("ELECTION_DATE"), err)
	}
	election := &Election{
		Name:         electionName,
		ElectionDate: electionDate,
	}
	db.FirstOrCreate(&election, election)
	return election, nil
}
