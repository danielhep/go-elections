package main

import (
	"sort"
	"time"

	"github.com/danielhep/go-elections/internal"
)

func sortCandidatesByLatestVotes(candidates []internal.BallotResponse) {
	sort.Slice(candidates, func(i, j int) bool {
		latestVotesI := getLatestVotes(candidates[i])
		latestVotesJ := getLatestVotes(candidates[j])
		return latestVotesI > latestVotesJ
	})
}

func getLatestVotes(candidate internal.BallotResponse) int {
	if len(candidate.VoteTallies) == 0 {
		return 0
	}

	latestUpdate := time.Time{}
	latestVotes := 0

	for _, tally := range candidate.VoteTallies {
		if tally.Update.Timestamp.After(latestUpdate) {
			latestUpdate = tally.Update.Timestamp
			latestVotes = tally.Votes
		}
	}

	return latestVotes
}
