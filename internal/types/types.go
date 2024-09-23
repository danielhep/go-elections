package types

import (
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

func (JurisdictionType) GormDataType() string {
	return "string"
}

// Structs to represent the data in DB
type Election struct {
	gorm.Model
	Name         string
	ElectionDate time.Time
	Contests     []Contest
	Updates      []Update
	Candidates   []Candidate
}

type Contest struct {
	gorm.Model
	Name          string
	District      string
	Jurisdictions pq.StringArray `gorm:"type:text[]"`
	Candidates    []Candidate
	ElectionID    uint
	Election      Election
}

type Candidate struct {
	gorm.Model
	Name        string
	Party       *string
	ContestID   uint
	Contest     Contest
	VoteTallies []VoteTally
	ElectionID  uint
	Election    Election
}

type Update struct {
	gorm.Model
	Timestamp        time.Time
	Hash             string
	JurisdictionType JurisdictionType
	VoteTallies      []VoteTally
	ElectionID       uint
	Election         Election
}

type VoteTally struct {
	gorm.Model
	CandidateID    uint
	Candidate      Candidate
	UpdateID       uint
	Contest        Contest
	ContestID      uint
	Update         Update
	Votes          int
	VotePercentage float32
}
