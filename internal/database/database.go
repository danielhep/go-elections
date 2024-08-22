package database

import (
	"fmt"
	"log"
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
	return csv.LoadCandidates(db.DB, data)
}

func (db *DB) UpdateVoteTallies(data []types.GenericVoteRecord, jurisdictionType types.JurisdictionType, hash string, timestamp time.Time) error {
	return csv.UpdateVoteTallies(db.DB, data, jurisdictionType, hash, timestamp)
}

func (db *DB) CheckAndProcessUpdate(data []types.GenericVoteRecord, hash string, jurisdictionType types.JurisdictionType) error {
	return csv.CheckAndProcessUpdate(db.DB, data, hash, jurisdictionType)
}
