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
				Name:     "dir",
				Aliases:  []string{"d"},
				Usage:    "Directory path containing CSV files",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "db",
				Usage:   "PostgreSQL database URL",
				EnvVars: []string{"PG_URL"},
			},
			&cli.StringFlag{
				Name:     "type",
				Aliases:  []string{"t"},
				Usage:    "Jurisdiction type (state or county)",
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
	dirPath := c.String("dir")
	dbURL := c.String("db")
	jurisdictionType := c.String("type")

	// Parse jurisdiction type
	var jType types.JurisdictionType
	switch strings.ToLower(jurisdictionType) {
	case "state":
		jType = types.StateJurisdiction
	case "county":
		jType = types.CountyJurisdiction
	default:
		return fmt.Errorf("invalid jurisdiction type: %s. Must be 'state' or 'county'", jurisdictionType)
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

	// Process CSV files
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			fmt.Printf("Processing file: %s\n", file.Name())

			// Extract date from filename
			datePart := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			datePart = strings.TrimPrefix(datePart, "webresults-")
			date, err := time.Parse("20060102", datePart)
			if err != nil {
				log.Printf("Failed to parse date from filename %s: %v", file.Name(), err)
				continue
			}

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

			// Load the candidates
			err = db.LoadCandidates(records)
			if err != nil {
				log.Printf("Failed to load candidates: %v", err)
			}

			// Update vote tallies
			err = db.UpdateVoteTallies(records, jType, hash, date)
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
