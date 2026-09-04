package service

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/digitalpapyrus/backend/internal/model"
)

// SendDraftSubmissionEmail sends an email to the customer and admin when a draft is submitted.
func (s *BookService) SendDraftSubmissionEmail(book *model.Book) error {
	user, err := s.userRepo.FindByID(book.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	order, err := s.orderRepo.FindByID(book.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	domain := "gmail.com"
	parts := strings.Split(smtpUsername, "@")
	if len(parts) > 1 {
		domain = parts[1]
	}
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
	dateStr := time.Now().Format(time.RFC1123Z)

	customerEmail := user.Email
	adminEmail := "digitalpapyrus15@gmail.com"

	var msgBuilder strings.Builder
	msgBuilder.WriteString("From: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", customerEmail))
	msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", adminEmail))
	msgBuilder.WriteString("Reply-To: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Subject: Draft - Pengajuan Draf Naskah Diterima - %s\r\n", book.Title))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")

	body := DraftSubmissionEmailTemplate
	body = strings.Replace(body, "INVOICE_PLACEHOLDER", order.Invoice, 1)
	body = strings.Replace(body, "CUSTOMER_NAME_PLACEHOLDER", user.Name, 1)
	body = strings.Replace(body, "BOOK_TITLE_PLACEHOLDER", book.Title, 1)
	body = strings.Replace(body, "BOOK_AUTHOR_PLACEHOLDER", book.Author, 1)
	body = strings.Replace(body, "BOOK_FORMAT_PLACEHOLDER", book.Format, 1)
	msgBuilder.WriteString(body)

	msg := []byte(msgBuilder.String())
	to := []string{customerEmail, adminEmail}
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// In test mode, we might not want to dial real SMTP if we don't have real credentials
	if smtpHost != "" && smtpPort != "" {
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			return fmt.Errorf("failed to send draft email: %w", err)
		}
	} else {
		// Mock sending for dev/testing when config is empty
		fmt.Printf("MOCK EMAIL SENT to %s and %s: Draft - Pengajuan Draf Naskah Diterima - %s\n", customerEmail, adminEmail, book.Title)
	}

	return nil
}

// SendValidationApproveEmail sends an email when a draft is approved.
func (s *BookService) SendValidationApproveEmail(book *model.Book) error {
	user, err := s.userRepo.FindByID(book.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	order, err := s.orderRepo.FindByID(book.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	domain := "gmail.com"
	parts := strings.Split(smtpUsername, "@")
	if len(parts) > 1 {
		domain = parts[1]
	}
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
	dateStr := time.Now().Format(time.RFC1123Z)

	customerEmail := user.Email
	adminEmail := "digitalpapyrus15@gmail.com"

	var msgBuilder strings.Builder
	msgBuilder.WriteString("From: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", customerEmail))
	msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", adminEmail))
	msgBuilder.WriteString("Reply-To: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Subject: Draft - Draf Naskah Disetujui - %s\r\n", book.Title))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")

	body := ValidationApproveEmailTemplate
	body = strings.Replace(body, "INVOICE_PLACEHOLDER", order.Invoice, 1)
	body = strings.Replace(body, "CUSTOMER_NAME_PLACEHOLDER", user.Name, 1)
	body = strings.Replace(body, "BOOK_TITLE_PLACEHOLDER", book.Title, 1)
	body = strings.Replace(body, "BOOK_AUTHOR_PLACEHOLDER", book.Author, 1)
	body = strings.Replace(body, "BOOK_FORMAT_PLACEHOLDER", book.Format, 1)
	msgBuilder.WriteString(body)

	msg := []byte(msgBuilder.String())
	to := []string{customerEmail, adminEmail}
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	if smtpHost != "" && smtpPort != "" {
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			return fmt.Errorf("failed to send approval email: %w", err)
		}
	} else {
		fmt.Printf("MOCK EMAIL SENT to %s and %s: Draft - Draf Naskah Disetujui - %s\n", customerEmail, adminEmail, book.Title)
	}

	return nil
}

// SendValidationRejectEmail sends an email when a draft is rejected.
func (s *BookService) SendValidationRejectEmail(book *model.Book) error {
	user, err := s.userRepo.FindByID(book.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	order, err := s.orderRepo.FindByID(book.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	domain := "gmail.com"
	parts := strings.Split(smtpUsername, "@")
	if len(parts) > 1 {
		domain = parts[1]
	}
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
	dateStr := time.Now().Format(time.RFC1123Z)

	customerEmail := user.Email
	adminEmail := "digitalpapyrus15@gmail.com"

	var msgBuilder strings.Builder
	msgBuilder.WriteString("From: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", customerEmail))
	msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", adminEmail))
	msgBuilder.WriteString("Reply-To: Digital Papyrus <supportdigitalpapyrus@gmail.com>\r\n")
	msgBuilder.WriteString(fmt.Sprintf("Subject: Draft - Draf Naskah Ditolak - %s\r\n", book.Title))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")

	body := ValidationRejectEmailTemplate
	body = strings.Replace(body, "INVOICE_PLACEHOLDER", order.Invoice, 1)
	body = strings.Replace(body, "CUSTOMER_NAME_PLACEHOLDER", user.Name, 1)
	body = strings.Replace(body, "BOOK_TITLE_PLACEHOLDER", book.Title, 1)
	body = strings.Replace(body, "BOOK_AUTHOR_PLACEHOLDER", book.Author, 1)
	body = strings.Replace(body, "BOOK_FORMAT_PLACEHOLDER", book.Format, 1)
	
	notes := book.Notes
	if notes == "" {
		notes = "Tidak ada catatan."
	}
	body = strings.Replace(body, "ADMIN_NOTES_PLACEHOLDER", notes, 1)
	msgBuilder.WriteString(body)

	msg := []byte(msgBuilder.String())
	to := []string{customerEmail, adminEmail}
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	if smtpHost != "" && smtpPort != "" {
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			return fmt.Errorf("failed to send rejection email: %w", err)
		}
	} else {
		fmt.Printf("MOCK EMAIL SENT to %s and %s: Draft - Draf Naskah Ditolak - %s\n", customerEmail, adminEmail, book.Title)
	}

	return nil
}
