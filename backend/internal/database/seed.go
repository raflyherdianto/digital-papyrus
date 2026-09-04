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
	if err := seedCoreServices(db); err != nil {
		return err
	}
	if err := seedOrders(db); err != nil {
		return err
	}
	if err := seedReviews(db); err != nil {
		return err
	}
	if err := seedRegions(db, cfg); err != nil {
		log.Printf("[DB Warning] seedRegions: %v", err)
	}
	if err := seedSettings(db); err != nil {
		log.Printf("[DB Warning] seedSettings: %v", err)
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
		`INSERT INTO users (id, email, password_hash, name, role, is_active, phone_number, address, province, city, zip_code) VALUES ($1, $2, $3, $4, 'superadmin', 1, '', '', '', '', '')`,
		uuid.New().String(), cfg.Seed.SuperAdminEmail, string(hash), cfg.Seed.SuperAdminName,
	)
	if err != nil {
		return fmt.Errorf("seed: insert superadmin: %w", err)
	}
	log.Printf("[DB] Seeded superadmin: %s", cfg.Seed.SuperAdminEmail)
	return nil
}

func seedDemoUsers(db *sql.DB, cfg *config.Config) error {
	demoUsers := []struct {
		Email    string
		Password string
		Name     string
		Role     string
	}{
		{cfg.Seed.DemoAuthorEmail, cfg.Seed.DemoAuthorPassword, cfg.Seed.DemoAuthorName, "author"},
		{cfg.Seed.DemoCustomerEmail, cfg.Seed.DemoCustomerPassword, cfg.Seed.DemoCustomerName, "customer"},
	}

	for _, u := range demoUsers {
		var existingCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", u.Email).Scan(&existingCount); err != nil {
			return fmt.Errorf("seed: check demo user %s: %w", u.Email, err)
		}
		if existingCount > 0 {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), cfg.Security.BcryptCost)
		if err != nil {
			return fmt.Errorf("seed: hash demo user password: %w", err)
		}
		_, err = db.Exec(
			`INSERT INTO users (id, email, password_hash, name, role, is_active, phone_number, address, province, city, zip_code) VALUES ($1, $2, $3, $4, $5, 1, '', '', '', '', '')`,
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

	categoryNames := []string{"Fiksi Kontemporer", "Filosofi Modern", "Teknologi & Sains", "Seni & Desain", "Psikologi", "Sastra Puisi"}
	categoryMap := make(map[string]string)

	for _, name := range categoryNames {
		id := uuid.New().String()
		slug := slugifyCategoryName(name)
		_, err := db.Exec(`INSERT INTO categories (id, name, slug) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, id, name, slug)
		if err != nil {
			return fmt.Errorf("seed: insert category %s: %w", name, err)
		}
		// fetch ID if skipped on conflict
		var catID string
		_ = db.QueryRow(`SELECT id FROM categories WHERE name = $1`, name).Scan(&catID)
		categoryMap[name] = catID
	}

	books := []struct {
		ID              string
		Title           string
		Author          string
		ISBN            string
		Badge           string
		GGKEY           string
		QRCBN           string
		Price           int
		Rating          float64
		ReviewCount     int
		Description     string
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
			Title: "TRANSFORMASI PENDIDIKAN DIGITAL: PERSPEKTIF PSIKOLOGI", Author: "Asrofi, S.Pd., M.Pd. – Cornelius Riko Bagus Nugroho - dkk", ISBN: "",
			Badge: "Regular", GGKEY: "", QRCBN: "",
			Price: 38000, Rating: 5.0, ReviewCount: 142,
			Description: "Buku kumpulan artikel berjudul \"Transformasi Pendidikan Digital: Perspektif Psikologi\" merupakan sebuah karya interdisipliner yang mengeksplorasi integrasi prinsip-prinsip psikologis dalam dunia pendidikan dan ekosistem digital yang terus berkembang. Melalui pendekatan teori psikologi yang dihubungkan dengan praktik relevan, buku ini menyajikan strategi inovatif untuk menciptakan lingkungan belajar yang lebih efektif serta memberikan solusi dalam menghadapi berbagai tantangan era modern, seperti dampak teknologi terhadap motivasi, konsentrasi, hingga kesehatan mental individu dan masyarakat.",
			ImageURL:    "/uploads/62f0ae62-d019-4b8c-a069-295b18836083.webp",
			CategoryID:  categoryMap["Filosofi Modern"], Status: "published", Stock: 60,
			Publisher: "Digital Papyrus Press", PublicationDate: "2025-10-12",
			Pages: 268, Format: "E-Book", Language: "Indonesia",
			Dimensions: "6.14 x 9.21 inches", Weight: "1.2",
		},
		{
			Title: "INTERVENSI PSIKOLOGI DALAM MEDIA SOSIAL DAN DIGITAL", Author: "Asrofi, S.Pd., M.Pd. – Cintania Amanda Putri – dkk", ISBN: "",
			Badge: "Regular", GGKEY: "", QRCBN: "",
			Price: 43000, Rating: 0.0, ReviewCount: 0,
			Description: "Buku berjudul \"Intervensi Psikologi Dalam Media Sosial Dan Digital\" merupakan karya kolektif mahasiswa Fakultas Psikologi Universitas Merdeka Malang yang mengulas dinamika kesehatan mental di era digital melalui berbagai perspektif psikologi. Melalui metode tinjauan literatur, buku ini mendalami fenomena modern seperti Fear of Missing Out (FOMO), cyberbullying, kecanduan media sosial, hingga penerapan psikodrama digital sebagai solusi terapi yang inovatif. Selain mengidentifikasi tantangan seperti keterbatasan interaksi fisik dan risiko perbandingan sosial, buku ini juga menawarkan wawasan strategis mengenai peluang teknologi dalam memperluas aksesibilitas intervensi psikologis bagi remaja dan dewasa muda.",
			ImageURL:    "/uploads/504aefec-1457-479b-9058-a153e99471d8.webp",
			CategoryID:  categoryMap["Filosofi Modern"], Status: "published", Stock: 100,
			Publisher: "Digital Papyrus", PublicationDate: "2026-04-23",
			Pages: 298, Format: "E-Book", Language: "Indonesia",
			Dimensions: "3 cm", Weight: "0.5",
		},
		{
			Title: "INTERAKSI PSIKOLOGI PERAN BUDAYA DAN BISNIS DI ERA DIGITAL", Author: "Asrofi, S.Pd., M.Pd – Paolo Constantine Julian Arnott Hungan – dkk", ISBN: "",
			Badge: "Regular", GGKEY: "", QRCBN: "",
			Price: 32000, Rating: 0.0, ReviewCount: 0,
			Description: "Buku berjudul Interaksi Psikologi: Peran Kebudayaan dan Bisnis di Era Digital merupakan karya ilmiah kolaboratif mahasiswa Fakultas Psikologi Universitas Merdeka Malang yang mengkaji transformasi interaksi manusia dari ruang fisik ke ruang virtual akibat perkembangan teknologi digital. Melalui pendekatan kritis dan empiris, buku ini mengulas beragam isu aktual seperti perilaku konsumen digital, dinamika identitas di media sosial, hingga implikasi psikologis dalam praktik bisnis and pendidikan, dengan tujuan memberikan pemahaman komprehensif mengenai relasi antara individu, budaya, dan bisnis di era modern.",
			ImageURL:    "/uploads/ccad1a69-2b7b-4604-810a-f79198704551.webp",
			CategoryID:  categoryMap["Filosofi Modern"], Status: "published", Stock: 80,
			Publisher: "Digital Papyrus", PublicationDate: "2026-04-28",
			Pages: 239, Format: "E-Book", Language: "Indonesia",
			Dimensions: "2", Weight: "0.2",
		},
		{
			Title: "INTERVENSI PSIKOLOGI DALAM MEDIA SOSIAL DAN DIGITAL V2", Author: "Asrofi, S.Pd., M.Pd. – Cintania Amanda Putri – dkk", ISBN: "",
			Badge: "Regular", GGKEY: "", QRCBN: "",
			Price: 34000, Rating: 0.0, ReviewCount: 0,
			Description: "Buku ini merupakan karya kolaboratif yang menghimpun pemikiran kritis dan hasil penelitian mendalam mengenai dinamika integrasi teknologi digital dalam pilar pendidikan, sistem informasi, dan perkembangan bahasa di Indonesia. Melalui berbagai perspektif, pembaca diajak mengeksplorasi peran strategis sistem informasi dalam memodernisasi tata kelola lembaga pendidikan, efektivitas media pembelajaran berbasis teknologi seperti Google Classroom, hingga inovasi kecerdasan buatan (machine learning) dalam aplikasi penerjemah bahasa isyarat (BISINDO). Selain aspek teknis, buku ini juga menelaah dampak sosial-budaya di era digital, termasuk pengaruh media sosial terhadap pergeseran gaya bahasa remaja serta pentingnya mempertimbangkan nilai-nilai budaya dalam pengembangan sistem informasi yang inklusif. Dengan menyajikan analisis komprehensif, karya ini bertujuan menjadi referensi penting bagi praktisi akademik dan masyarakat umum dalam menghadapi tantangan serta memanfaatkan peluang transformasi digital demi kemajuan pendidikan yang berkelanjutan.",
			ImageURL:    "/uploads/56e938b7-0f1c-4b26-a15d-389f7cc831cc.webp",
			CategoryID:  categoryMap["Fiksi Kontemporer"], Status: "published", Stock: 35,
			Publisher: "Digital Papyrus", PublicationDate: "2026-05-06",
			Pages: 354, Format: "E-Book", Language: "Indonesia",
			Dimensions: "14 x 21 cm", Weight: "0.2",
		},
		{
			ID: "b898e661-b365-4b34-9690-c4c0bc241d38",
			Title: "TITAH KETENTRAMAN MENUJU SANG PENCIPTA", Author: "Zainul Arifin", ISBN: "",
			Badge: "Regular", GGKEY: "", QRCBN: "",
			Price: 42000, Rating: 0.0, ReviewCount: 0,
			Description: "\"Titah Ketentraman Menuju Sang Pencipta\" adalah sebuah karya sastra berbentuk kumpulan puisi (antologi) yang mendalam, lahir dari refleksi spiritual dan kontemplasi tajam sang penulis, Zainul Arifin. Ditulis di Pasuruan sepanjang rentang waktu beberapa tahun, buku ini menyajikan untaian kata yang sarat akan makna kehidupan, kerinduan ilahiah, serta kritik sosial yang dibalut dengan bahasa yang menyentuh jiwa",
			ImageURL:    "/uploads/e525396b-73a8-49af-8641-ebd2ba9f9852.webp",
			CategoryID:  categoryMap["Sastra Puisi"], Status: "published", Stock: 20,
			Publisher: "Digital Papyrus", PublicationDate: "2026-06-11",
			Pages: 2016, Format: "E-Book", Language: "Indonesia",
			Dimensions: "2.4 cm", Weight: "100",
		},
	}

	stmt, err := db.Prepare(`INSERT INTO books (
id, title, author, isbn, badge, ggkey, qrcbn, price, rating, review_count,
description, image_url, category_id, status, stock,
publisher, publication_date, pages, format, language, dimensions, weight
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
ON CONFLICT (id) DO UPDATE SET
title=EXCLUDED.title, author=EXCLUDED.author, price=EXCLUDED.price, description=EXCLUDED.description,
image_url=EXCLUDED.image_url, category_id=EXCLUDED.category_id, status=EXCLUDED.status, stock=EXCLUDED.stock,
publisher=EXCLUDED.publisher, publication_date=EXCLUDED.publication_date, pages=EXCLUDED.pages,
format=EXCLUDED.format, language=EXCLUDED.language, dimensions=EXCLUDED.dimensions, weight=EXCLUDED.weight`)
	if err != nil {
		return fmt.Errorf("seed: prepare books insert: %w", err)
	}
	defer stmt.Close()

	for _, b := range books {
		var isbn interface{} = b.ISBN
		if b.ISBN == "" || b.ISBN == "-" {
			isbn = nil
		}

		id := b.ID
		if id == "" {
			id = uuid.New().String()
		}

		_, err := stmt.Exec(
			id, b.Title, b.Author, isbn, b.Badge, b.GGKEY, b.QRCBN,
			b.Price, b.Rating, b.ReviewCount,
			b.Description, b.ImageURL, b.CategoryID, b.Status, b.Stock,
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
			Title: "Paket Basic", Description: "Paket starter untuk penerbitan buku digital dengan desain cover dasar dan legalitas terbitan.",
			Icon: "auto_stories", Tier: "basic", Price: 500000, PriceLabel: "Rp 500k",
			Features:   []string{"Penerbitan Digital (PDF)", "Desain Kover Standar", "Maksimal 150 Hal", "Surat Bukti Proses (LOA)*", "Surat Bukti Terbit*", "Sertifikat Penulis*", "Royalti Penjualan", "Upload ke Repository", "Template Buku", "Diskon HAKI Rp. 25.000", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Starter", SortOrder: 1,
		},
		{
			Title: "Paket Silver", Description: "Paket menengah untuk penerbitan buku fisik dengan desain kover dan layout naskah profesional.",
			Icon: "workspace_premium", Tier: "silver", Price: 750000, PriceLabel: "Rp 750k",
			Features:   []string{"Penerbitan Buku Cetak", "Layout & Desain Kover", "Buku untuk Penulis 2(A5)/1(B5)", "Surat Bukti Proses (LOA)*", "Surat Bukti Terbit*", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Sertifikat Penulis*", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "Diskon HaKI Rp. 50.000", "GRATIS Ongkos Kirim*", "Maksimal 150 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Menengah", SortOrder: 2,
		},
		{
			Title: "Paket Gold", Description: "Paket lengkap untuk penerbitan profesional dengan legalitas penuh, mockup 3D, dan buku fisik berlimpah.",
			Icon: "military_tech", Tier: "gold", Price: 1000000, PriceLabel: "Rp 1 Jt",
			Features:   []string{"Penerbitan Profesional", "Buku Penulis (5 A5 / 4 B5)", "Desain Mockup 3D", "Layout Naskah Buku", "Desain Kover Eksklusif", "Surat Bukti Proses (LOA)*", "Surat Bukti Terbit*", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Preview e-Book (PDF)", "Sertifikat Penulis", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "Diskon HaKI Rp. 75.000", "GRATIS Ongkos Kirim*", "Maksimal 200 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: true, Badge: "Profesional", SortOrder: 3,
		},
		{
			Title: "Paket Platinum", Description: "Paket eksklusif untuk penulis premium dengan fasilitas terlengkap dan jumlah buku fisik terbanyak.",
			Icon: "diamond", Tier: "platinum", Price: 1500000, PriceLabel: "Rp 1.5 Jt",
			Features:   []string{"Penerbitan Premium", "Buku Penulis (10 A5 / 8 B5)", "Diskon HaKI Rp 100.000", "Layout Naskah Buku", "Desain Kover Eksklusif", "Surat Bukti Proses (LOA)*", "Surat Bukti Terbit*", "2 Buku Arsip Perpusnas", "1 Buku Arsip Perpusda", "1 Buku Arsip Penerbit", "Full e-Book (PDF)", "Preview e-Book (PDF)", "Sertifikat Penulis", "Royalti Penjualan", "Upload ke Repository", "GRATIS Packing Buku", "Template Buku", "Mockup 3D*", "Laminasi Doff atau Glossy", "Book Paper/HVS", "Wrapping Buku", "GRATIS Ongkos Kirim*", "Maksimal 250 Hal", "Ukuran A5, B5 (Unesco/Reguler)"},
			IsFeatured: false, Badge: "Eksklusif", SortOrder: 4,
		},
	}

	stmt, err := db.Prepare(`INSERT INTO services (
id, title, description, icon, tier, price, price_label, features,
is_featured, badge, sort_order, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1)`)
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

func seedCoreServices(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM core_services").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check core_services: %w", err)
	}
	if count > 0 {
		return nil
	}

	type cSvc struct {
		Title       string
		Description string
		Icon        string
		SortOrder   int
	}

	coreServices := []cSvc{
		{
			Title: "Layout & Desain Kover", Description: "Tata letak interior naskah yang rapi dan rancangan desain sampul paling menarik, disesuaikan dengan permintaan serta tren visual buku eksklusif.",
			Icon: "design_services", SortOrder: 1,
		},
		{
			Title: "Penerbitan & Legalitas", Description: "Proses penerbitan buku yang terintegrasi dengan standar legalitas nasional untuk memastikan karya Anda terdaftar secara resmi dan terlindungi.",
			Icon: "gavel", SortOrder: 2,
		},
		{
			Title: "Cetak Satuan & Print on Demand", Description: "Dari satu eksemplar hingga cetak oplah besar menggunakan kertas terbaik (Book Paper/HVS) dengan opsi Softcover atau Hardcover eksekutif (Doff/Glossy).",
			Icon: "print", SortOrder: 3,
		},
	}

	stmt, err := db.Prepare(`INSERT INTO core_services (
id, title, description, icon, sort_order, is_active
) VALUES ($1, $2, $3, $4, $5, 1)`)
	if err != nil {
		return fmt.Errorf("seed: prepare core_services insert: %w", err)
	}
	defer stmt.Close()

	for _, s := range coreServices {
		_, err = stmt.Exec(
			uuid.New().String(), s.Title, s.Description, s.Icon, s.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("seed: insert core_service %s: %w", s.Title, err)
		}
	}
	log.Printf("[DB] Seeded %d core services", len(coreServices))
	return nil
}

func seedOrders(db *sql.DB) error {
	var customerID string
	err := db.QueryRow("SELECT id FROM users WHERE email = 'customer@digitalpapyrus.web.id' LIMIT 1").Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[DB] Warning: customer user not found, skipping order seeding")
			return nil
		}
		return fmt.Errorf("seed: get customer user for orders: %w", err)
	}

	var book1ID, book2ID string
	_ = db.QueryRow("SELECT id FROM books ORDER BY title LIMIT 1").Scan(&book1ID)
	_ = db.QueryRow("SELECT id FROM books ORDER BY title LIMIT 1 OFFSET 1").Scan(&book2ID)

	var service1ID string
	_ = db.QueryRow("SELECT id FROM services ORDER BY title LIMIT 1").Scan(&service1ID)
	orders := []struct {
		Invoice         string
		Notes           string
		TotalQty        int
		TotalWeight     int
		TotalPrice      int
		PaymentType     string
		Status          string
		ShippingName    string
		ShippingService string
		ShippingPrice   int
		Details         []struct {
			ServiceID  string
			BookID     string
			Qty        int
			TotalPrice int
		}
	}{
		{
			Invoice: "INV-20260507001", Notes: "Demo order pending", TotalQty: 1, TotalWeight: 500, TotalPrice: 185000,
			PaymentType: "Bank Transfer", Status: "pending", ShippingName: "JNE", ShippingService: "REG", ShippingPrice: 20000,
			Details: []struct {
				ServiceID  string
				BookID     string
				Qty        int
				TotalPrice int
			}{{ServiceID: "", BookID: book1ID, Qty: 1, TotalPrice: 185000}},
		},
		{
			Invoice: "INV-20260507002", Notes: "Demo order confirmed", TotalQty: 2, TotalWeight: 1000, TotalPrice: 430000,
			PaymentType: "Bank Transfer", Status: "confirmed", ShippingName: "JNE", ShippingService: "REG", ShippingPrice: 20000,
			Details: []struct {
				ServiceID  string
				BookID     string
				Qty        int
				TotalPrice int
			}{{ServiceID: "", BookID: book1ID, Qty: 1, TotalPrice: 185000}, {ServiceID: "", BookID: book2ID, Qty: 1, TotalPrice: 245000}},
		},
		{
			Invoice: "INV-20260507003", Notes: "Demo order shipped", TotalQty: 1, TotalWeight: 0, TotalPrice: 500000,
			PaymentType: "Credit Card", Status: "shipped", ShippingName: "Internal", ShippingService: "Express", ShippingPrice: 0,
			Details: []struct {
				ServiceID  string
				BookID     string
				Qty        int
				TotalPrice int
			}{{ServiceID: service1ID, BookID: "", Qty: 1, TotalPrice: 500000}},
		},
		{
			Invoice: "INV-20260507004", Notes: "Demo order delivered", TotalQty: 2, TotalWeight: 500, TotalPrice: 685000,
			PaymentType: "Virtual Account", Status: "delivered", ShippingName: "J&T", ShippingService: "YES", ShippingPrice: 25000,
			Details: []struct {
				ServiceID  string
				BookID     string
				Qty        int
				TotalPrice int
			}{{ServiceID: service1ID, BookID: "", Qty: 1, TotalPrice: 500000}, {ServiceID: "", BookID: book1ID, Qty: 1, TotalPrice: 185000}},
		},
		{
			Invoice: "INV-20260507005", Notes: "Demo order cancelled", TotalQty: 1, TotalWeight: 500, TotalPrice: 245000,
			PaymentType: "Bank Transfer", Status: "cancelled", ShippingName: "JNE", ShippingService: "REG", ShippingPrice: 20000,
			Details: []struct {
				ServiceID  string
				BookID     string
				Qty        int
				TotalPrice int
			}{{ServiceID: "", BookID: book2ID, Qty: 1, TotalPrice: 245000}},
		},
	}

	for _, order := range orders {
		var existing string
		err := db.QueryRow("SELECT id FROM orders WHERE invoice = $1 LIMIT 1", order.Invoice).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("seed: check order %s: %w", order.Invoice, err)
		}
		if existing != "" {
			continue
		}

		orderID := uuid.New().String()
		_, err = db.Exec(`INSERT INTO orders (id, invoice, user_id, notes, total_qty, total_weight, total_price, payment_type, status, shipping_name, shipping_service, shipping_price)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			orderID, order.Invoice, customerID, order.Notes, order.TotalQty, order.TotalWeight, order.TotalPrice, order.PaymentType, order.Status, order.ShippingName, order.ShippingService, order.ShippingPrice)
		if err != nil {
			return fmt.Errorf("seed: insert order %s: %w", order.Invoice, err)
		}

		for _, detail := range order.Details {
			_, err = db.Exec(`INSERT INTO order_details (id, order_id, service_id, book_id, qty, total_price)
						  VALUES ($1, $2, $3, $4, $5, $6)`,
				uuid.New().String(), orderID, detail.ServiceID, detail.BookID, detail.Qty, detail.TotalPrice)
			if err != nil {
				return fmt.Errorf("seed: insert order_details %s: %w", order.Invoice, err)
			}
		}
	}

	log.Printf("[DB] Seeded orders")
	return nil
}

func seedReviews(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM reviews").Scan(&count)
	if err != nil {
		return fmt.Errorf("seed: check reviews: %w", err)
	}
	if count > 0 {
		return nil
	}

	var customerID string
	err = db.QueryRow("SELECT id FROM users WHERE email = 'customer@digitalpapyrus.web.id' LIMIT 1").Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[DB] Warning: customer user not found, skipping review seeding")
			return nil
		}
		return fmt.Errorf("seed: get customer user: %w", err)
	}

	var orderID string
	err = db.QueryRow("SELECT id FROM orders WHERE user_id = $1 AND status IN ('delivered', 'completed') LIMIT 1", customerID).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[DB] Warning: delivered/completed order not found, skipping review seeding")
			return nil
		}
		return fmt.Errorf("seed: get delivered/completed order: %w", err)
	}

	var book1ID, book2ID string
	_ = db.QueryRow("SELECT id FROM books ORDER BY title LIMIT 1").Scan(&book1ID)
	_ = db.QueryRow("SELECT id FROM books ORDER BY title LIMIT 1 OFFSET 1").Scan(&book2ID)

	var service1ID string
	_ = db.QueryRow("SELECT id FROM services ORDER BY title LIMIT 1").Scan(&service1ID)

	serviceIDsJSON := fmt.Sprintf(`["%s"]`, service1ID)
	bookIDsJSON := fmt.Sprintf(`["%s", "%s"]`, book1ID, book2ID)
	detailsJSON := fmt.Sprintf(`{"service_%s": "layanan cepat", "book_%s": "buku bagus", "book_%s": "kualitas cetak baik"}`, service1ID, book1ID, book2ID)
	ratingJSON := fmt.Sprintf(`{"service_%s": 5, "book_%s": 5, "book_%s": 4}`, service1ID, book1ID, book2ID)

	_, err = db.Exec(`INSERT INTO reviews (id, user_id, order_id, service_id, book_id, details, rating) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New().String(), customerID, orderID, serviceIDsJSON, bookIDsJSON, detailsJSON, ratingJSON)

	if err != nil {
		return fmt.Errorf("seed: insert review: %w", err)
	}

	log.Printf("[DB] Seeded reviews")
	return nil
}

func seedSettings(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM app_settings").Scan(&count); err == nil && count > 0 {
		return nil
	}

	settings := map[string]string{
		"tax":                 "11",
		"service_fee":         "5000",
		"discount":            "0",
		"origin_phone":        "+6281234567890",
		"origin_address":      "Jl. Raya Darmo No. 45, Darmo, Kec. Wonokromo",
		"origin_province":     "35",
		"origin_city":         "3578",
		"origin_district":     "357804",
		"origin_village_code": "3578041005",
		"origin_zip_code":     "60241",
		"bank_name":            "Bank BCA",
		"bank_account_number":  "8735091234",
		"bank_account_holder":  "PT Digital Papyrus Indonesia",
	}

	for k, v := range settings {
		_, err := db.Exec(`
			INSERT INTO app_settings (key, value, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
		`, k, v)
		if err != nil {
			return fmt.Errorf("seed app_settings %s: %w", k, err)
		}
	}
	log.Printf("[DB] Seeded default app settings")
	return nil
}
