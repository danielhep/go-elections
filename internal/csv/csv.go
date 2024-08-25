package csv

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/danielhep/go-elections/internal/types"
	"github.com/gocarina/gocsv"
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

// Function to process state-level data
func ProcessContests(records []types.GenericVoteRecord) ([]types.Contest, error) {
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
