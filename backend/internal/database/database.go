// Package database provides SQLite connection management and schema migration.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (CGo-free)
)

// New creates a new SQLite database connection with production-grade settings.
func New(dbPath string) (*sql.DB, error) {
	// Ensure the data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("database: failed to create data directory %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-20000&_foreign_keys=ON&_txlock=immediate", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open %s: %w", dbPath, err)
	}

	// Connection pool settings for SQLite
	db.SetMaxOpenConns(1) // SQLite supports only one writer at a time
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	log.Printf("[DB] Connected to SQLite: %s", dbPath)
	return db, nil
}

// Migrate creates all tables if they do not already exist.
func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id          TEXT PRIMARY KEY,
			email       TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name        TEXT NOT NULL,
			role        TEXT NOT NULL CHECK(role IN ('superadmin','author','customer')),
			is_active   INTEGER NOT NULL DEFAULT 1,
			phone_number TEXT DEFAULT '',
			address     TEXT DEFAULT '',
			province    TEXT DEFAULT '',
			city        TEXT DEFAULT '',
			zip_code    TEXT DEFAULT '',
			created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
		);`,

		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);`,

		`CREATE TABLE IF NOT EXISTS categories (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL UNIQUE,
			slug        TEXT NOT NULL UNIQUE,
			created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
		);`,

		`CREATE TABLE IF NOT EXISTS books (
			id               TEXT PRIMARY KEY,
			title            TEXT NOT NULL,
			author           TEXT NOT NULL,
			isbn             TEXT UNIQUE,
			badge            TEXT DEFAULT 'Regular',
			ggkey            TEXT DEFAULT '',
			qrcbn            TEXT DEFAULT '',
			price            INTEGER NOT NULL DEFAULT 0,
			rating           REAL NOT NULL DEFAULT 0,
			review_count     INTEGER NOT NULL DEFAULT 0,
			description      TEXT DEFAULT '',
			synopsis         TEXT DEFAULT '',
			image_url        TEXT DEFAULT '',
			category_id      TEXT,
			status           TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published','archived')),
			stock            INTEGER NOT NULL DEFAULT 0,
			publisher        TEXT DEFAULT '',
			publication_date TEXT DEFAULT '',
			pages            INTEGER DEFAULT 0,
			format           TEXT DEFAULT '',
			language         TEXT DEFAULT '',
			dimensions       TEXT DEFAULT '',
			weight           TEXT DEFAULT '',
			created_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_books_status ON books(status);`,
		`CREATE INDEX IF NOT EXISTS idx_books_category_id ON books(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);`,

		`CREATE TABLE IF NOT EXISTS services (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			icon        TEXT DEFAULT '',
			tier        TEXT NOT NULL CHECK(tier IN ('basic','silver','gold','platinum')),
			price       INTEGER NOT NULL DEFAULT 0,
			price_label TEXT DEFAULT '',
			features    TEXT DEFAULT '[]',
			is_featured INTEGER NOT NULL DEFAULT 0,
			badge       TEXT DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			is_active   INTEGER NOT NULL DEFAULT 1,
			created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
		);`,

		`CREATE INDEX IF NOT EXISTS idx_services_tier ON services(tier);`,
		`CREATE INDEX IF NOT EXISTS idx_services_sort_order ON services(sort_order);`,

		`CREATE TABLE IF NOT EXISTS core_services (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			icon        TEXT DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			is_active   INTEGER NOT NULL DEFAULT 1,
			created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
		);`,

		`CREATE INDEX IF NOT EXISTS idx_core_services_sort_order ON core_services(sort_order);`,

		`CREATE TABLE IF NOT EXISTS orders (
			id               TEXT PRIMARY KEY,
			invoice          TEXT NOT NULL UNIQUE,
			user_id          TEXT NOT NULL,
			notes            TEXT DEFAULT '',
			total_qty        INTEGER NOT NULL DEFAULT 0,
			total_weight     INTEGER NOT NULL DEFAULT 0,
			total_price      INTEGER NOT NULL DEFAULT 0,
			payment_type     TEXT DEFAULT '',
			status           TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','confirmed','shipped','delivered','cancelled','completed')),
			shipping_name    TEXT DEFAULT '',
			shipping_service TEXT DEFAULT '',
			shipping_price   INTEGER NOT NULL DEFAULT 0,
			created_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);`,

		`CREATE TABLE IF NOT EXISTS order_details (
			id          TEXT PRIMARY KEY,
			order_id    TEXT NOT NULL,
			service_id  TEXT DEFAULT '',
			book_id     TEXT DEFAULT '',
			qty         INTEGER NOT NULL DEFAULT 1,
			total_price INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (order_id) REFERENCES orders(id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_order_details_order_id ON order_details(order_id);`,

		`CREATE TABLE IF NOT EXISTS reviews (
			id          TEXT PRIMARY KEY,
			user_id     TEXT NOT NULL,
			order_id    TEXT NOT NULL,
			service_id  TEXT DEFAULT '[]',
			book_id     TEXT DEFAULT '[]',
			details     TEXT DEFAULT '{}',
			rating      TEXT DEFAULT '{}',
			created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at  DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (order_id) REFERENCES orders(id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_order_id ON reviews(order_id);`,
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("database: begin migration tx: %w", err)
	}

	for _, m := range migrations {
		if _, err := tx.Exec(m); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("database: migration failed: %w\nSQL: %s", err, m)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: commit migration tx: %w", err)
	}

	if err := ensureCategorySlugs(db); err != nil {
		return fmt.Errorf("database: ensure category slugs: %w", err)
	}

	if err := ensureUserColumns(db); err != nil {
		return fmt.Errorf("database: ensure user columns: %w", err)
	}

	if err := ensureBookColumns(db); err != nil {
		return fmt.Errorf("database: ensure book columns: %w", err)
	}

	if err := ensureOrderStatuses(db); err != nil {
		return fmt.Errorf("database: ensure order statuses: %w", err)
	}

	log.Println("[DB] Schema migration completed successfully")
	return nil
}

func ensureUserColumns(db *sql.DB) error {
	columns := []string{"phone_number", "address", "province", "city", "zip_code"}
	for _, col := range columns {
		if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE users ADD COLUMN %s TEXT DEFAULT ''`, col)); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
				return fmt.Errorf("add %s column to users: %w", col, err)
			}
		}
	}
	return nil
}

func ensureBookColumns(db *sql.DB) error {
	columns := map[string]string{
		"badge": "TEXT DEFAULT 'Regular'",
		"ggkey": "TEXT DEFAULT ''",
		"qrcbn": "TEXT DEFAULT ''",
	}
	for col, def := range columns {
		if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE books ADD COLUMN %s %s`, col, def)); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
				return fmt.Errorf("add %s column to books: %w", col, err)
			}
		}
	}
	return nil
}

func ensureOrderStatuses(db *sql.DB) error {
	var createSQL string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'orders'`).Scan(&createSQL); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("inspect orders table: %w", err)
	}

	if strings.Contains(createSQL, "confirmed") && strings.Contains(createSQL, "delivered") {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("rebuild orders table: begin: %w", err)
	}
	defer tx.Rollback()

	statements := []string{
		`PRAGMA foreign_keys = OFF`,
		`ALTER TABLE orders RENAME TO orders_legacy`,
		`CREATE TABLE orders (
			id               TEXT PRIMARY KEY,
			invoice          TEXT NOT NULL UNIQUE,
			user_id          TEXT NOT NULL,
			notes            TEXT DEFAULT '',
			total_qty        INTEGER NOT NULL DEFAULT 0,
			total_weight     INTEGER NOT NULL DEFAULT 0,
			total_price      INTEGER NOT NULL DEFAULT 0,
			payment_type     TEXT DEFAULT '',
			status           TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','confirmed','shipped','delivered','cancelled','completed')),
			shipping_name    TEXT DEFAULT '',
			shipping_service TEXT DEFAULT '',
			shipping_price   INTEGER NOT NULL DEFAULT 0,
			created_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at       DATETIME NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`INSERT INTO orders (id, invoice, user_id, notes, total_qty, total_weight, total_price, payment_type, status, shipping_name, shipping_service, shipping_price, created_at, updated_at)
		 SELECT id, invoice, user_id, notes, total_qty, total_weight, total_price, payment_type,
			CASE status
				WHEN 'processing' THEN 'confirmed'
				WHEN 'completed' THEN 'delivered'
				ELSE status
			END,
			shipping_name, shipping_service, shipping_price, created_at, updated_at
		 FROM orders_legacy`,
		`DROP TABLE orders_legacy`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`PRAGMA foreign_keys = ON`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("rebuild orders table: %w\nSQL: %s", err, stmt)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("rebuild orders table: commit: %w", err)
	}

	return nil
}

func ensureCategorySlugs(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE categories ADD COLUMN slug TEXT`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("add slug column: %w", err)
		}
	}

	rows, err := db.Query(`SELECT id, name FROM categories ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return fmt.Errorf("select categories for slug backfill: %w", err)
	}
	defer rows.Close()

	type categoryRow struct {
		ID   string
		Name string
	}

	var categories []categoryRow
	for rows.Next() {
		var c categoryRow
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return fmt.Errorf("scan category for slug backfill: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate categories for slug backfill: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin slug backfill tx: %w", err)
	}

	used := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		base := slugifyCategoryName(c.Name)
		slug := base
		idx := 2
		for {
			if _, exists := used[slug]; !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", base, idx)
			idx++
		}
		used[slug] = struct{}{}

		if _, err := tx.Exec(`UPDATE categories SET slug = ? WHERE id = ?`, slug, c.ID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("update category slug: %w", err)
		}
	}

	if _, err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)`); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create slug index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit slug backfill tx: %w", err)
	}

	return nil
}

func slugifyCategoryName(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, slug)

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "category"
	}
	return slug
}
