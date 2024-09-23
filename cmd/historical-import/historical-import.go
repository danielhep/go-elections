package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielhep/go-elections/internal/csv"
	"github.com/danielhep/go-elections/internal/database"
	"github.com/danielhep/go-elections/internal/types"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "historical-import",
		Usage: "Import historical election data from CSV files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db",
				Usage:    "PostgreSQL database URL",
				EnvVars:  []string{"PG_URL"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "date",
				Usage:    "Election date (YYYY-MM-DD)",
				Aliases:  []string{"d"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name of the election (2024 Primary)",
				Aliases:  []string{"n"},
				Required: true,
			},
		},
		Action: runImport,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runImport(c *cli.Context) error {
	dirPath := c.Args().Get(0)
	if dirPath == "" {
		return fmt.Errorf("directory path is required")
	}
	dbURL := c.String("db")
	electionName := c.String("name")
	electionDateStr := c.String("date")
	electionDate, err := time.Parse("2006-01-02", electionDateStr)
	if err != nil {
		return fmt.Errorf("failed to parse election date: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB(dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Migrate schema
	if err := db.MigrateSchema(); err != nil {
		return fmt.Errorf("failed to migrate schema: %v", err)
	}

	// Create an election object
	election := types.Election{
		Name:         electionName,
		ElectionDate: electionDate,
	}
	db.FirstOrCreate(&election, election)

	// Process CSV files
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			fmt.Printf("Processing file: %s\n", file.Name())

			// Determine jurisdiction type
			var jType types.JurisdictionType
			if strings.Contains(file.Name(), "allstate") {
				jType = types.StateJurisdiction
			} else if strings.Contains(file.Name(), "webresults") {
				jType = types.CountyJurisdiction
			} else {
				return fmt.Errorf("unknown jurisdiction type from filename: %s", file.Name())
			}

			fmt.Printf("Detected jurisdiction type: %s\n", jType)

			// Extract date from filename
			datePart := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			datePart = strings.TrimSuffix(datePart, "_allstate")
			datePart = strings.TrimSuffix(datePart, "-final")
			datePart = strings.TrimPrefix(datePart, "webresults-")
			date, err := time.Parse("20060102", datePart)
			if err != nil {
				log.Printf("Failed to parse date from filename %s: %v", file.Name(), err)
				continue
			}

			fmt.Printf("Detected date: %s\n", date)

			// Open and parse CSV file
			f, err := os.Open(filepath.Join(dirPath, file.Name()))
			if err != nil {
				log.Printf("Failed to open file %s: %v", file.Name(), err)
				continue
			}
			defer f.Close()

			records, hash, err := csv.Parse(f, jType)
			if err != nil {
				log.Printf("Failed to parse CSV file %s: %v", file.Name(), err)
				continue
			}

			// Load the responses
			err = db.LoadBallotResponses(records, election)
			if err != nil {
				log.Printf("Failed to load ballot responses: %v", err)
				continue
			}

			// Update vote tallies
			err = db.UpdateVoteTallies(records, hash, date, election)
			if err != nil {
				log.Printf("Failed to update vote tallies for file %s: %v", file.Name(), err)
				continue
			}

			fmt.Printf("Successfully processed file: %s\n", file.Name())
		}
	}

	fmt.Println("Historical data import completed.")
	return nil
}
