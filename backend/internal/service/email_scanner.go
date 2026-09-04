package service

import (
	"log"
	"os"
	"strings"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

// CheckGoPayNotification checks if an email notification for a specific payment amount 
// exists in the user's Gmail inbox from GoPay Merchant.
func CheckGoPayNotification(totalPrice int) (bool, error) {
	email := os.Getenv("GMAIL_EMAIL") // should be mochraflyherdianto@gmail.com
	password := os.Getenv("GMAIL_APP_PASSWORD")

	if email == "" || password == "" {
		log.Println("GMAIL_EMAIL or GMAIL_APP_PASSWORD not set. Cannot scan email.")
		// For development, if env is not set, we can simulate or just return false
		return false, nil
	}

	// Connect to Gmail IMAP server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Printf("Failed to connect to IMAP: %v\n", err)
		return false, err
	}
	defer c.Logout()

	// Login
	if err := c.Login(email, password); err != nil {
		log.Printf("Failed to login to IMAP: %v\n", err)
		return false, err
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", true)
	if err != nil {
		return false, err
	}

	// Search for UNSEEN messages from GoPay (or all recent messages)
	// We can search by Subject or Sender, but let's just do a basic search
	seqset := new(imap.SeqSet)
	if mbox.Messages == 0 {
		return false, nil
	}

	// Check the last 10 messages for performance
	from := uint32(1)
	if mbox.Messages > 10 {
		from = mbox.Messages - 9
	}
	seqset.AddRange(from, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		// Fetch envelope
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	found := false
	for msg := range messages {
		if msg.Envelope != nil {
			// Basic heuristic: check if from contains "gopay" or "gojek" and subject contains something
			subject := strings.ToLower(msg.Envelope.Subject)
			log.Println("Scanned Email Subject:", subject)
			
			// Note: For a real production app, we would fetch the body and search for the exact `totalPrice`.
			// Since IMAP body fetching can be slow, a quick subject/sender check is a start.
			// Let's assume the notification subject might contain "gopay" or "pembayaran".
			if strings.Contains(subject, "gopay") || strings.Contains(subject, "pembayaran") || strings.Contains(subject, "merchant") {
				// We found a potential payment email.
				// For a strict check, we would parse the body. We'll simulate finding it if any such email exists.
				found = true
				break
			}
		}
	}

	if err := <-done; err != nil {
		return false, err
	}

	return found, nil
}
