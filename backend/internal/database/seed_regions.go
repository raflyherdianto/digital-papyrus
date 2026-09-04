package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/digitalpapyrus/backend/internal/config"
)

//go:embed regions_seed.sql
var regionSeedSQL string

func seedRegions(db *sql.DB, cfg *config.Config) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM provinces").Scan(&count); err == nil && count > 0 {
		return nil // already seeded
	}

	log.Println("[DB Seed] Seeding Indonesian regional data from local embedded SQL file...")
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("seed regions: begin tx: %w", err)
	}

	if _, err := tx.Exec(regionSeedSQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("seed regions: execute sql: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("seed regions: commit tx: %w", err)
	}

	log.Printf("[DB Seed] Successfully seeded regional data from local SQL in %v", time.Since(start))
	return nil
}
