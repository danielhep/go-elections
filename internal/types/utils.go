package types

import (
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