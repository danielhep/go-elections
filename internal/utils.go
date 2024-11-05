package internal

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Helper function to extract contest name and district
func extractContestInfo(raceField string) (contestName string, district string) {
	parts := strings.SplitN(raceField, " - ", 2)
	if len(parts) == 2 {
		district = normalizeString(parts[0])
		contestName = normalizeString(parts[1])
	} else {
		contestName = normalizeString(raceField)
		if strings.Contains(contestName, "United States") {
			district = "Federal"
		} else {
			district = "State of Washington"
		}
	}
	return
}

// Helper function to normalize string capitalization
func normalizeString(s string) string {
	s = strings.ReplaceAll(s, "Lt.", "Lieutenant")
	s = strings.ReplaceAll(s, "U.S.", "United States")
	s = strings.ReplaceAll(s, "#0", "")
	s = strings.ReplaceAll(s, "No. ", "")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "SUPREME COURT", "State Supreme Court")
	s = strings.ReplaceAll(s, "of the United States", "")
	s = strings.ReplaceAll(s, "STATEWIDE", "State of Washington")
	words := strings.Fields(s)
	for i, word := range words {
		lcWord := strings.ToLower(word)
		if lcWord != "of" &&
			lcWord != "the" &&
			lcWord != "and" &&
			lcWord != "in" &&
			lcWord != "for" &&
			word != "US" &&
			word != "USA" {
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

func getContestKey(ballotTitle string, districtName string) string {
	ballotTitle = strings.ReplaceAll(ballotTitle, " ", "_")
	districtName = strings.ReplaceAll(districtName, " ", "_")
	return fmt.Sprintf("%s-%s", ballotTitle, districtName)
}

func getCandidateKey(contestID uint, ballotResponse string) string {
	return fmt.Sprintf("%d-%s", contestID, ballotResponse)
}

func GetElectionKey(electionName string) string {
	return strings.ReplaceAll(strings.ToLower(electionName), " ", "_")
}
