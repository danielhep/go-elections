package internal

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// StateCSVRecord represents the structure of each row in the state CSV file
type StateCSVRecord struct {
	Race                   string  `csv:"Race"`
	Candidate              string  `csv:"Candidate"`
	Party                  string  `csv:"Party"`
	Votes                  int     `csv:"Votes"`
	PercentageOfTotalVotes float64 `csv:"PercentageOfTotalVotes"`
	JurisdictionName       string  `csv:"JurisdictionName"`
}

func (rec StateCSVRecord) ToGeneric() GenericVoteRecord {
	contestName, district := extractContestInfo(rec.Race)
	return GenericVoteRecord{
		DistrictName:     district,
		BallotTitle:      contestName,
		BallotResponse:   normalizeString(rec.Candidate),
		VotePercentage:   float32(rec.PercentageOfTotalVotes),
		Votes:            rec.Votes,
		PartyPreference:  extractParty(rec.Party),
		JurisdictionType: StateJurisdiction,
	}
}

// CountyCSVRecord represents the structure of each row in the county CSV file
type CountyCSVRecord struct {
	GEMSContestID               string  `csv:"GEMS Contest ID"`
	ContestSortSeq              int     `csv:"Contest Sort Seq"`
	DistrictType                string  `csv:"District Type"`
	DistrictTypeSubheading      string  `csv:"District Type Subheading"`
	DistrictName                string  `csv:"District Name"`
	BallotTitle                 string  `csv:"Ballot Title"`
	BallotsCountedForDistrict   int     `csv:"Ballots Counted for District"`
	RegisteredVotersForDistrict int     `csv:"Registered Voters for District"`
	PercentTurnoutForDistrict   float64 `csv:"Percent Turnout for District"`
	CandidateSortSeq            int     `csv:"Candidate Sort Seq"`
	BallotResponse              string  `csv:"Ballot Response"`
	PartyPreference             string  `csv:"Party Preference"`
	Votes                       int     `csv:"Votes"`
	PercentOfVotes              float64 `csv:"Percent of Votes"`
}

func (rec CountyCSVRecord) ToGeneric() GenericVoteRecord {
	return GenericVoteRecord{
		DistrictName:     normalizeString(rec.DistrictName),
		BallotTitle:      normalizeString(rec.BallotTitle),
		BallotResponse:   normalizeString(rec.BallotResponse),
		VotePercentage:   float32(rec.PercentOfVotes),
		Votes:            rec.Votes,
		PartyPreference:  extractParty(rec.PartyPreference),
		JurisdictionType: CountyJurisdiction,
	}
}

type GenericVoteRecord struct {
	DistrictName     string
	BallotTitle      string
	BallotResponse   string
	Votes            int
	VotePercentage   float32
	PartyPreference  string
	JurisdictionType JurisdictionType
}

type JurisdictionType string

const (
	StateJurisdiction  JurisdictionType = "State"
	CountyJurisdiction JurisdictionType = "County"
)

// Structs to represent the data in DB
type Election struct {
	gorm.Model
	ID           string
	Name         string
	ElectionDate time.Time
	Contests     []Contest
	Updates      []Update
	Candidates   []BallotResponse
}

type Contest struct {
	gorm.Model
	BallotTitle     string
	District        string
	ContestKey      string         `gorm:"index"`
	Jurisdictions   pq.StringArray `gorm:"type:text[]"`
	BallotResponses []BallotResponse
	ElectionID      string
	Election        Election `gorm:"constraint:OnDelete:CASCADE,onUpdate:CASCADE"`
}

type BallotResponse struct {
	gorm.Model
	Name        string
	Party       *string
	ContestID   uint
	Contest     Contest `gorm:"constraint:OnDelete:CASCADE,onUpdate:CASCADE"`
	VoteTallies []VoteTally
	ElectionID  string
	Election    Election `gorm:"constraint:OnDelete:CASCADE,onUpdate:CASCADE"`
}

type Update struct {
	gorm.Model
	Timestamp        time.Time
	Hash             string `gorm:"uniqueIndex"`
	JurisdictionType JurisdictionType
	VoteTallies      []VoteTally
	ElectionID       string
	Election         Election
}

func (u *Update) BeforeDelete(tx *gorm.DB) (err error) {
	// Delete all vote tallies associated with this update
	fmt.Printf("Deleting vote tallies for update %v\n", u.ID)
	if err := tx.Unscoped().Where("update_id = ?", u.ID).Delete(&VoteTally{}).Error; err != nil {
		return fmt.Errorf("error deleting vote tallies: %v", err)
	}
	return nil
}

type VoteTally struct {
	gorm.Model
	BallotResponseID uint
	BallotResponse   BallotResponse `gorm:"constraint:OnDelete:CASCADE,onUpdate:CASCADE"`
	UpdateID         uint
	Contest          Contest `gorm:"constraint:OnDelete:CASCADE,onUpdate:CASCADE"`
	ContestID        uint
	Update           Update `gorm:"constraint:OnDelete:CASCADE"`
	Votes            int
	VotePercentage   float32
}
