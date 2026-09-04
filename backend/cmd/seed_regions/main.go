package main

import (
	"log"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.New(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully!")
}
