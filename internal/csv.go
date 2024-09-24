package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gocarina/gocsv"
)

func Parse(reader io.ReadCloser, jurisdictionType JurisdictionType) ([]GenericVoteRecord, string, error) {
	// Create a TeeReader to read the body and calculate hash simultaneously
	hashReader := sha256.New()
	teeReader := io.TeeReader(reader, hashReader)

	var records []GenericVoteRecord
	switch jurisdictionType {
	case StateJurisdiction:
		var stateRecords []*StateCSVRecord
		if err := gocsv.Unmarshal(teeReader, &stateRecords); err != nil {
			return nil, "", err
		}
		for _, record := range stateRecords {
			records = append(records, record.ToGeneric())
		}
	case CountyJurisdiction:
		var countyRecords []*CountyCSVRecord
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
func ParseFromURL(url string, jurisdictionType JurisdictionType) ([]GenericVoteRecord, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	return Parse(resp.Body, jurisdictionType)
}

// Function to process state-level data
func ProcessContests(records []GenericVoteRecord, election Election) ([]Contest, error) {
	contestMap := make(map[string]*Contest)

	for _, record := range records {
		// Create or get Contest
		contestKey := getContestKey(record.BallotTitle, record.DistrictName)
		contest, exists := contestMap[contestKey]
		if !exists {
			contest = &Contest{
				BallotTitle:     record.BallotTitle,
				District:        record.DistrictName,
				ContestKey:      contestKey,
				BallotResponses: []BallotResponse{},
				ElectionID:      election.ID,
			}
			contestMap[contestKey] = contest
		}

		// Create ballot response and add to Contest
		ballotResponse := BallotResponse{
			Name:       record.BallotResponse,
			Party:      &record.PartyPreference,
			ElectionID: election.ID,
		}
		contest.BallotResponses = append(contest.BallotResponses, ballotResponse)
	}
	// Convert map to slice
	contests := make([]Contest, 0, len(contestMap))
	for _, contest := range contestMap {
		contests = append(contests, *contest)
	}

	return contests, nil
}
