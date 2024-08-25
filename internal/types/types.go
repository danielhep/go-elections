package types

import (
	"time"

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
	PartyPreference  string
	JurisdictionType JurisdictionType
}

type JurisdictionType string

const (
	StateJurisdiction  JurisdictionType = "State"
	CountyJurisdiction JurisdictionType = "County"
)

// Structs to represent the data in DB
type Contest struct {
	gorm.Model
	Name       string
	District   string
	Candidates []Candidate
}

type Candidate struct {
	gorm.Model
	Name        string
	Party       *string
	ContestID   uint
	Contest     Contest
	VoteTallies []VoteTally
}

type Update struct {
	gorm.Model
	Timestamp        time.Time
	Hash             string
	JurisdictionType JurisdictionType
	VoteTallies      []VoteTally
}

type VoteTally struct {
	gorm.Model
	CandidateID uint
	Candidate   Candidate
	UpdateID    uint
	Contest     Contest
	ContestID   uint
	Update      Update
	Votes       int
}
