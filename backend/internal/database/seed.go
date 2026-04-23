package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/digitalpapyrus/backend/internal/config"
)

// Seed populates the database with initial data including superadmin and dummy content
// matching the frontend's catalog and promo sections.
func Seed(db *sql.DB, cfg *config.Config) error {
	if err := seedSuperAdmin(db, cfg); err != nil {
		return err
	}
	if err := seedDemoUsers(db, cfg); err != nil {
		return err
	}
	if err := seedCategoriesAndBooks(db); err != nil {
		return err
	}
	if err := seedServices(db); err != nil {
		return err
	}
	log.Println("[DB] Data seeding completed successfully")
	return nil
}

func seedSuperAdmin(db *sql.DB, cfg *config.Config) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'superadmin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check superadmin: %w", err)
	}
	if count > 0 {
		return nil // already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Seed.SuperAdminPassword), cfg.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("seed: hash superadmin password: %w", err)
	}

	_, err = db.Exec(
		`INSERT INTO users (id, email, password_hash, name, role, is_active) VALUES (?, ?, ?, ?, 'superadmin', 1)`,
		uuid.New().String(), cfg.Seed.SuperAdminEmail, string(hash), cfg.Seed.SuperAdminName,
	)
	if err != nil {
		return fmt.Errorf("seed: insert superadmin: %w", err)
	}
	log.Printf("[DB] Seeded superadmin: %s", cfg.Seed.SuperAdminEmail)
	return nil
}

func seedDemoUsers(db *sql.DB, cfg *config.Config) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'author'").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check author: %w", err)
	}
	if count > 0 {
		return nil
	}

	demoUsers := []struct {
		Email string
		Name  string
		Role  string
	}{
		{"author@digitalpapyrus.web.id", "Demo Author", "author"},
		{"customer@digitalpapyrus.web.id", "Demo Customer", "customer"},
	}

	for _, u := range demoUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte("Demo@2026!"), cfg.Security.BcryptCost)
		if err != nil {
			return fmt.Errorf("seed: hash demo user password: %w", err)
		}
		_, err = db.Exec(
			`INSERT INTO users (id, email, password_hash, name, role, is_active) VALUES (?, ?, ?, ?, ?, 1)`,
			uuid.New().String(), u.Email, string(hash), u.Name, u.Role,
		)
		if err != nil {
			return fmt.Errorf("seed: insert demo user %s: %w", u.Email, err)
		}
		log.Printf("[DB] Seeded demo user: %s (%s)", u.Email, u.Role)
	}
	return nil
}

func seedCategoriesAndBooks(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check books: %w", err)
	}
	if count > 0 {
		return nil
	}

	categoryNames := []string{"Fiksi Kontemporer", "Filosofi Modern", "Teknologi & Sains", "Seni & Desain"}
	categoryMap := make(map[string]string)

	for _, name := range categoryNames {
		id := uuid.New().String()
		slug := slugifyCategoryName(name)
		_, err := db.Exec(`INSERT INTO categories (id, name, slug) VALUES (?, ?, ?)`, id, name, slug)
		if err != nil {
			return fmt.Errorf("seed: insert category %s: %w", name, err)
		}
		categoryMap[name] = id
	}

	books := []struct {
		Title           string
		Author          string
		ISBN            string
		Price           int
		Rating          float64
		ReviewCount     int
		Description     string
		Synopsis        string
		ImageURL        string
		CategoryID      string
		Status          string
		Stock           int
		Publisher       string
		PublicationDate string
		Pages           int
		Format          string
		Language        string
		Dimensions      string
		Weight          string
	}{
		{
			Title: "The Silent Echo", Author: "Evelyn Vance", ISBN: "978-3-16-148410-1",
			Price: 185000, Rating: 4.8, ReviewCount: 96,
			Description: "A haunting exploration of memory and silence in modern society.",
			Synopsis:    "An evocative novel that traces the journey of a sound archivist as she discovers that the world's most powerful echoes are not heard, but felt.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuBAjri-Yg7PVc-Tbks3rSUsyHyEOPo3nimS4QfdQXgMI6FqFGtTwn-uzHAyfeEV-zDRvY51LbxLgweIfQwX2RSmkry5pxlQJ04eIWZx43lh1POj2o16r98WAXY6j1b4IYKEes1p4a7nqstGvPk1WQCmfek5O05GAitwDxNixobb4QbnqAqWSAf30hLxS33GCClbu8NMk8FaezqLrf_98OkXhfn2i7qNabf1rQBC_zhe_fMgNgCBK3DMLmy4o4rOqdbbB5QMeS6z_F4",
			CategoryID:  categoryMap["Fiksi Kontemporer"], Status: "published", Stock: 120,
			Publisher: "Digital Papyrus Press", PublicationDate: "2026-01-15",
			Pages: 284, Format: "Softcover", Language: "Indonesian",
			Dimensions: "14 x 21 cm", Weight: "0.5",
		},
		{
			Title: "Digital Ontologies", Author: "Dr. Marcus Thorne", ISBN: "978-3-16-148410-2",
			Price: 245000, Rating: 4.9, ReviewCount: 128,
			Description: "A groundbreaking analysis of being and consciousness in the digital age.",
			Synopsis:    "Dr. Thorne explores how technology reshapes our understanding of existence, identity, and the nature of reality itself in an increasingly digitized world.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuAg44EpYgx4EqdLV2VJriLVBNaFeuCRnkh-So0dQtU96rWLEndO66bZA31N-brnXwK6lMiV6s_bJs28eDKQNR3Zo2I4Npx661MUOM43jhvlEbF33BULvsWPHDbfPj0g00bdL_FiL7301yG2WsG3N1cPkF9whQwfNcfBJpz1hkVg6IewAoPbnnpYvoonA3nqAbjBqYNzz_Z_No8ITRPbhUAwgGisluW6LcLoGrtyTl7MlJen5dTcWyc9SxxZje708Scp076hQiYFu7c",
			CategoryID:  categoryMap["Filosofi Modern"], Status: "published", Stock: 85,
			Publisher: "Digital Papyrus Press", PublicationDate: "2025-11-20",
			Pages: 412, Format: "Hardcover XL", Language: "English (UK)",
			Dimensions: "6.14 x 9.21 inches", Weight: "1.2",
		},
		{
			Title: "The Last Algorithm", Author: "Sarah Jenkins", ISBN: "978-3-16-148410-3",
			Price: 120000, Rating: 4.5, ReviewCount: 64,
			Description: "A thriller set in the world of artificial intelligence and ethical computing.",
			Synopsis:    "When the world's most advanced AI begins making decisions no one predicted, one programmer must race against time to find the last algorithm.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuBQb7XBj1l03NM7mRXEgQXMgFhDCy7l-vHVol32dx0qE3s7G390fLHU55FKbf3CoTDc4oUmXt3UsJv1ivxQ_9tQk6TF2GcjSVC-JvzMn3XN0tb42_UbBFFFhOclCiXjmU67x3MmMu_1XT2uqECW5rpSK9BpQIZjwLts1_uSRsTnHyNXKtOkg4tSXV9GLSWuI_mNJ48kzdabqjZw4UC_HKhSzzL_fdugiEKydLjA1HbpUdHTI710SlAD983OOijhEk-NK7qEDvD_MR8",
			CategoryID:  categoryMap["Teknologi & Sains"], Status: "published", Stock: 200,
			Publisher: "Digital Papyrus Press", PublicationDate: "2026-03-01",
			Pages: 198, Format: "Softcover", Language: "Indonesian",
			Dimensions: "14 x 21 cm", Weight: "0.28",
		},
		{
			Title: "Urban Melodies", Author: "Julian Grey", ISBN: "978-3-16-148410-4",
			Price: 155000, Rating: 4.7, ReviewCount: 82,
			Description: "Poetry and prose intertwine in this love letter to city life.",
			Synopsis:    "Julian Grey captures the rhythm of urban existence through lyrical prose, blending music, architecture, and human connection into a symphony of words.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuB7xUT7LtjWws2X6Nhypxp3Qg-XQAvzj-_nuDVWkHp_WCh18uYKBq7RpOL-LPSk-GBBEO2Rf2LMnNkjqZi0xIWP0M6zKHa6eaRhX1QShqeVy7xTI-VjwP-1UaqXczUNXOir2GBtGUgk4zf1J_UAWtD-A-SnD7BgQ_6f-8ZJUuQekdTXDAsBCW8uoS5HTojalV6IpWS_KQR4vqVoChFAYXDEKCtV2b6mQ8_JNwMrcTuLnKPfja2f-X5wERMQ8pgpFvd-g9IGXIarTNg",
			CategoryID:  categoryMap["Seni & Desain"], Status: "published", Stock: 150,
			Publisher: "Digital Papyrus Press", PublicationDate: "2025-08-10",
			Pages: 256, Format: "Softcover", Language: "Indonesian",
			Dimensions: "14 x 21 cm", Weight: "0.32",
		},
		{
			Title: "Infinite Canvas", Author: "Liam O'Connell", ISBN: "978-3-16-148410-5",
			Price: 210000, Rating: 4.9, ReviewCount: 115,
			Description: "Where art meets technology - a stunning visual essay on digital creativity.",
			Synopsis:    "A coffee-table book that investigates the limitless possibilities of digital art, featuring interviews with leading digital artists and immersive visuals.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuBK92hK5POEH69Wyp0HUJDgQovmXFfKx4VYFoeuaU4_EexKcWfI0_-9JsTQ3H4A1YD6TvF2KZfokdS3LJG5ldAr451CP2X9BSTEDtgyT_tpUAKsrw3MakCpHdokRoS-sm_LTaMiBolnaOodvMIEqUXmIRX0o6wqiqGmZosw78Jx7wCOUKcS2PEL__O5HfZqaLZxsOkf-8WaYAs_bTKzbISoeeSnh3xIswsx124xjX1vpqgCM-_8UJ_goNZaywcyTVEPRi7u8qn3zHU",
			CategoryID:  categoryMap["Seni & Desain"], Status: "published", Stock: 75,
			Publisher: "Digital Papyrus Press", PublicationDate: "2026-02-14",
			Pages: 320, Format: "Hardcover XL", Language: "English (UK)",
			Dimensions: "8.5 x 11 inches", Weight: "2.1",
		},
		{
			Title: "Architecture of Thought", Author: "Sofia Rossi", ISBN: "978-3-16-148410-6",
			Price: 275000, Rating: 5.0, ReviewCount: 142,
			Description: "A magnum opus exploring the intersection of architecture and philosophy.",
			Synopsis:    "Sofia Rossi masterfully crafts a narrative that is part philosophical inquiry and part poetic meditation, making this one of the most anticipated releases in contemporary editorial literature.",
			ImageURL:    "https://lh3.googleusercontent.com/aida-public/AB6AXuAp8zd6LtCFVVJ7J-SlNW7yNxLnqkImIiM8ltkm75XuS4zOHd-L1_yt3nhr4eYd6Ql0wFbd9j3sviyZdqRxKFkQ66QXVO31mQ6HdBg4UkDgxo427QauEjnbxODPcr54sK8b2NnRgdulR6idvtp35AjfJiDLvt7mPI8EAE6FXhgLqu3xOnpYlh65eQjeXGUXqgcSWqMQu_wSQF31Smf5vbx1CdtyKfSA4iciUaT6WAHam1MO8JSish77mefIl-6yd5ETk6wwwRGtaz8",
			CategoryID:  categoryMap["Filosofi Modern"], Status: "published", Stock: 60,
			Publisher: "Digital Papyrus Press", PublicationDate: "2025-10-12",
			Pages: 342, Format: "Hardcover XL", Language: "English (UK)",
			Dimensions: "6.14 x 9.21 inches", Weight: "1.2",
		},
	}

	stmt, err := db.Prepare(`INSERT INTO books (
id, title, author, isbn, price, rating, review_count,
description, synopsis, image_url, category_id, status, stock,
publisher, publication_date, pages, format, language, dimensions, weight
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("seed: prepare books insert: %w", err)
	}
	defer stmt.Close()

	for _, b := range books {
		_, err := stmt.Exec(
			uuid.New().String(), b.Title, b.Author, b.ISBN,
			b.Price, b.Rating, b.ReviewCount,
			b.Description, b.Synopsis, b.ImageURL, b.CategoryID, b.Status, b.Stock,
			b.Publisher, b.PublicationDate, b.Pages, b.Format, b.Language, b.Dimensions, b.Weight,
		)
		if err != nil {
			return fmt.Errorf("seed: insert book %s: %w", b.Title, err)
		}
	}
	log.Printf("[DB] Seeded %d categories and %d books", len(categoryNames), len(books))
	return nil
}

func seedServices(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check services: %w", err)
	}
	if count > 0 {
		return nil
	}

	type svc struct {
		Title       string
		Description string
		Icon        string
		Tier        string
		Price       int
		PriceLabel  string
		Features    []string
		IsFeatured  bool
		Badge       string
		SortOrder   int
	}

	services := []svc{
		{
			Title: "Paket Basic", Description: "Paket starter untuk penerbitan buku dengan E-ISBN dan desain cover dasar.",
			Icon: "auto_stories", Tier: "basic", Price: 275000, PriceLabel: "Rp 275k",
			Features:   []string{"E-ISBN Perpusnas RI", "Cover Design", "Surat bukti proses/LOA*", "Surat Bukti Terbit*", "Full e-Book (PDF)", "Sertifikat Penulis*", "Royalti Penjualan", "Upload ke Repository", "Template Buku", "Diskon HAKI Rp. 25.000", "Maksimal 150 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Starter", SortOrder: 1,
		},
		{
			Title: "Paket Silver", Description: "Paket menengah dengan ISBN cetak, layout naskah, dan buku fisik untuk penulis.",
			Icon: "workspace_premium", Tier: "silver", Price: 455000, PriceLabel: "Rp 455k",
			Features:   []string{"ISBN Perpusnas RI", "Layout Naskah Buku", "Cover Design", "Surat bukti proses/LOA*", "Surat Bukti Terbit*", "Buku untuk Penulis (2 buku A5 atau 1 buku B5)", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Sertifikat Penulis*", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "Diskon HaKI Rp. 50.000", "GRATIS Ongkos Kirim*", "Maksimal 150 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Menengah", SortOrder: 2,
		},
		{
			Title: "Paket Gold", Description: "Paket profesional dengan ISBN, DOI, mockup 3D, dan buku fisik berlimpah.",
			Icon: "military_tech", Tier: "gold", Price: 765000, PriceLabel: "Rp 765k",
			Features:   []string{"ISBN Perpusnas RI", "Digital Object Identifier (DOI)", "Layout Naskah Buku", "Cover Design", "Surat bukti proses/LOA*", "Surat Bukti Terbit*", "Buku untuk Penulis (5 buku A5 atau 4 buku B5)", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Preview e-Book (PDF)", "Sertifikat Penulis", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Mockup 3D*", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "Diskon HaKI Rp. 75.000", "GRATIS Ongkos Kirim*", "Maksimal 200 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: true, Badge: "Terpopuler", SortOrder: 3,
		},
		{
			Title: "Paket Platinum", Description: "Paket eksklusif premium dengan semua fitur dan buku fisik terbanyak.",
			Icon: "diamond", Tier: "platinum", Price: 965000, PriceLabel: "Rp 965k",
			Features:   []string{"ISBN Perpusnas RI", "Digital Object Identifier (DOI)", "Layout Naskah Buku", "Cover Design", "Surat bukti proses/LOA*", "Surat Bukti Terbit*", "Buku untuk Penulis (10 buku A5 atau 8 buku B5)", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Preview e-Book (PDF)", "Sertifikat Penulis", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Mockup 3D*", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "Diskon HaKI Rp. 100.000", "GRATIS Ongkos Kirim*", "Maksimal 250 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Eksklusif", SortOrder: 4,
		},
	}

	stmt, err := db.Prepare(`INSERT INTO services (
id, title, description, icon, tier, price, price_label, features,
is_featured, badge, sort_order, is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`)
	if err != nil {
		return fmt.Errorf("seed: prepare services insert: %w", err)
	}
	defer stmt.Close()

	for _, s := range services {
		featuresJSON, err := json.Marshal(s.Features)
		if err != nil {
			return fmt.Errorf("seed: marshal features for %s: %w", s.Title, err)
		}
		featured := 0
		if s.IsFeatured {
			featured = 1
		}
		_, err = stmt.Exec(
			uuid.New().String(), s.Title, s.Description, s.Icon, s.Tier,
			s.Price, s.PriceLabel, string(featuresJSON),
			featured, s.Badge, s.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("seed: insert service %s: %w", s.Title, err)
		}
	}
	log.Printf("[DB] Seeded %d services", len(services))
	return nil
}
