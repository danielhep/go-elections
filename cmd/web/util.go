package main

import (
	"sort"
	"time"

	"github.com/danielhep/go-elections/internal/types"
)

func sortCandidatesByLatestVotes(candidates []types.Candidate) {
	sort.Slice(candidates, func(i, j int) bool {
		latestVotesI := getLatestVotes(candidates[i])
		latestVotesJ := getLatestVotes(candidates[j])
		return latestVotesI > latestVotesJ
	})
}

func getLatestVotes(candidate types.Candidate) int {
	if len(candidate.VoteTallies) == 0 {
		return 0
	}

	latestUpdate := time.Time{}
	latestVotes := 0

	for _, tally := range candidate.VoteTallies {
		updateTime, err := time.Parse(time.RFC3339, tally.Update.Timestamp)
		if err != nil {
			continue // Skip this tally if we can't parse the timestamp
		}

		if updateTime.After(latestUpdate) {
			latestUpdate = updateTime
			latestVotes = tally.Votes
		}
	}

	return latestVotes
}
