package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GoPayCache struct {
	Token      string `json:"gopay_token"`
	MerchantID string `json:"gopay_merchant_id"`
}

type GoBizService struct {
	mu               sync.Mutex
	token            string
	merchantID       string
	cachePath        string
	lastLoginAttempt time.Time
	lastLoginError   error
}

var (
	gobizSvc     *GoBizService
	gobizSvcOnce sync.Once
)

// GetGoBizService returns the singleton instance of GoBizService.
func GetGoBizService() *GoBizService {
	gobizSvcOnce.Do(func() {
		gobizSvc = &GoBizService{
			cachePath: ".gopay_cache.json",
		}
	})
	return gobizSvc
}

// getAuthHeaders returns HTTP headers for GoBiz API.
func getAuthHeaders(uniqueID, token string) map[string]string {
	auth := "Bearer"
	if token != "" {
		auth = "Bearer " + token
	}
	return map[string]string{
		"Accept":              "application/json, text/plain, */*",
		"Accept-Language":     "id",
		"Authentication-Type": "go-id",
		"Authorization":       auth,
		"Connection":          "keep-alive",
		"Content-Type":        "application/json",
		"Gojek-Country-Code":  "ID",
		"Gojek-Timezone":      "Asia/Jakarta",
		"Origin":              "https://portal.gofoodmerchant.co.id",
		"Referer":             "https://portal.gofoodmerchant.co.id/",
		"Sec-Fetch-Dest":      "empty",
		"Sec-Fetch-Mode":      "cors",
		"Sec-Fetch-Site":      "cross-site",
		"User-Agent":          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",
		"X-AppVersion":        "platform-v3.107.0-94ce5d57",
		"X-PhoneMake":         "Windows 10 64-bit",
		"X-PhoneModel":        "Chrome 149.0.0.0 on Windows 10 64-bit",
		"X-Platform":          "Web",
		"X-User-Locale":       "en-US",
		"X-User-Type":         "merchant",
		"sec-ch-ua":           `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`,
		"sec-ch-ua-mobile":    "?0",
		"sec-ch-ua-platform":  `"Windows"`,
		"x-DeviceOS":          "Web",
		"x-appId":             "go-biz-web-dashboard",
		"x-uniqueid":          uniqueID,
	}
}

// loadCache loads token and merchant ID from local cache file.
func (s *GoBizService) loadCache() GoPayCache {
	var cache GoPayCache
	file, err := os.Open(s.cachePath)
	if err != nil {
		return cache
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	_ = dec.Decode(&cache)
	return cache
}

// saveCache writes token and merchant ID to local cache file.
func (s *GoBizService) saveCache(cache GoPayCache) {
	file, err := os.Create(s.cachePath)
	if err != nil {
		log.Printf("[GoBizService] Gagal menyimpan cache: %v\n", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cache)
}

// execCurl runs curl command to perform raw HTTP requests, bypassing Cloudflare/JA3 fingerprinting issues.
func execCurl(method string, urlStr string, headers map[string]string, body interface{}) ([]byte, int, error) {
	args := []string{"-4", "-s", "-w", "\n%{http_code}", "-X", method, urlStr}
	for k, v := range headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
	}
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal body: %w", err)
		}
		args = append(args, "--data-raw", string(bodyBytes))
	}

	cmd := exec.Command("curl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("curl execution failed: %w (stderr: %s)", err, stderr.String())
	}

	outStr := stdout.String()
	lines := strings.Split(strings.TrimRight(outStr, "\r\n"), "\n")
	if len(lines) < 2 {
		return nil, 0, fmt.Errorf("unexpected output format from curl: %q", outStr)
	}

	statusCodeStr := strings.TrimSpace(lines[len(lines)-1])
	var statusCode int
	_, err = fmt.Sscanf(statusCodeStr, "%d", &statusCode)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse HTTP status code %q: %w. Full output: %s", statusCodeStr, err, outStr)
	}

	bodyStr := strings.Join(lines[:len(lines)-1], "\n")
	return []byte(bodyStr), statusCode, nil
}

// isTokenValid checks if the GoBiz session token is still valid.
func (s *GoBizService) isTokenValid(token string) bool {
	uniqueID := uuid.New().String()
	body := map[string]interface{}{
		"from":    0,
		"to":      1,
		"_source": []string{"id"},
	}
	headers := getAuthHeaders(uniqueID, token)
	_, statusCode, err := execCurl("POST", "https://api.gobiz.co.id/v1/merchants/search", headers, body)
	if err != nil {
		return false
	}
	return statusCode != http.StatusUnauthorized
}

// doLogin performs the email/password login flow against GoBiz.
func (s *GoBizService) doLogin() error {
	s.lastLoginAttempt = time.Now()
	email := os.Getenv("GOPAY_EMAIL")
	password := os.Getenv("GOPAY_PASSWORD")

	if email == "" || password == "" {
		err := errors.New("GOPAY_EMAIL or GOPAY_PASSWORD not set in environment variables")
		s.lastLoginError = err
		return err
	}

	uniqueID := uuid.New().String()
	headers := getAuthHeaders(uniqueID, "")

	// 1. Request Login (triggers session registration)
	reqBody := map[string]interface{}{
		"email":      email,
		"login_type": "password",
		"client_id":  "go-biz-web-new",
	}
	respBytes, statusCode, err := execCurl("POST", "https://api.gobiz.co.id/goid/login/request", headers, reqBody)
	if err != nil {
		err = fmt.Errorf("failed to send login request: %w", err)
		s.lastLoginError = err
		return err
	}

	if statusCode < 200 || statusCode >= 300 {
		var errData struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		var loginErr error
		if err := json.Unmarshal(respBytes, &errData); err == nil && len(errData.Errors) > 0 {
			loginErr = fmt.Errorf("login request failed: %s (status %d)", errData.Errors[0].Message, statusCode)
		} else {
			loginErr = fmt.Errorf("login request failed with status %d: %s", statusCode, string(respBytes))
		}
		s.lastLoginError = loginErr
		return loginErr
	}

	// 2. Request Token
	tokenReqBody := map[string]interface{}{
		"client_id":  "go-biz-web-new",
		"grant_type": "password",
		"data": map[string]string{
			"email":    email,
			"password": password,
		},
	}
	tokenRespBytes, tokenStatusCode, err := execCurl("POST", "https://api.gobiz.co.id/goid/token", headers, tokenReqBody)
	if err != nil {
		err = fmt.Errorf("failed to send token request: %w", err)
		s.lastLoginError = err
		return err
	}

	if tokenStatusCode < 200 || tokenStatusCode >= 300 {
		loginErr := fmt.Errorf("login failed with status %d: %s", tokenStatusCode, string(tokenRespBytes))
		s.lastLoginError = loginErr
		return loginErr
	}

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Errors       []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(tokenRespBytes, &tokenData); err != nil {
		err = fmt.Errorf("failed to unmarshal token response: %w", err)
		s.lastLoginError = err
		return err
	}

	if len(tokenData.Errors) > 0 {
		loginErr := fmt.Errorf("login failed: %s", tokenData.Errors[0].Message)
		s.lastLoginError = loginErr
		return loginErr
	}

	s.token = tokenData.AccessToken
	s.lastLoginError = nil

	// Save token to cache
	cache := s.loadCache()
	cache.Token = s.token
	s.saveCache(cache)

	log.Printf("[GoBizService] Login berhasil untuk email %s. Token baru disimpan.\n", email)
	return nil
}

// getUserMerchants retrieves the list of merchants associated with the user account.
func (s *GoBizService) getUserMerchants() (interface{}, error) {
	uniqueID := uuid.New().String()
	body := map[string]interface{}{
		"from":    0,
		"to":      50,
		"_source": []string{"id", "merchant_name"},
	}
	headers := getAuthHeaders(uniqueID, s.token)
	respBytes, statusCode, err := execCurl("POST", "https://api.gobiz.co.id/v1/merchants/search", headers, body)
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("failed to get merchants status %d: %s", statusCode, string(respBytes))
	}

	var data interface{}
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// extractMerchantID resolves the first merchant ID from the GoBiz search list.
func extractMerchantID(raw interface{}) (string, error) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return "", errors.New("unexpected merchants response format")
	}

	// 1. Try directly checking the "merchants" field
	if list, exists := m["merchants"]; exists {
		if arr, ok := list.([]interface{}); ok && len(arr) > 0 {
			if first, ok := arr[0].(map[string]interface{}); ok {
				if id, ok := first["id"].(string); ok {
					return id, nil
				}
				if id, ok := first["merchant_id"].(string); ok {
					return id, nil
				}
			}
		}
	}

	// 2. Try checks inside "hits"
	if hits, exists := m["hits"]; exists {
		if hitsMap, ok := hits.(map[string]interface{}); ok {
			if hitsArr, ok := hitsMap["hits"].([]interface{}); ok && len(hitsArr) > 0 {
				if firstHit, ok := hitsArr[0].(map[string]interface{}); ok {
					if source, ok := firstHit["_source"].(map[string]interface{}); ok {
						if id, ok := source["id"].(string); ok {
							return id, nil
						}
						if id, ok := source["merchant_id"].(string); ok {
							return id, nil
						}
					}
				}
			}
		}
		if hitsArr, ok := hits.([]interface{}); ok && len(hitsArr) > 0 {
			if first, ok := hitsArr[0].(map[string]interface{}); ok {
				if id, ok := first["id"].(string); ok {
					return id, nil
				}
			}
		}
	}

	// 3. Try checks inside "data"
	if data, exists := m["data"]; exists {
		if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
			if first, ok := arr[0].(map[string]interface{}); ok {
				if id, ok := first["id"].(string); ok {
					return id, nil
				}
			}
		}
	}

	return "", errors.New("no merchants found in response")
}

// init ensures token validity and caches properties.
func (s *GoBizService) init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := s.loadCache()

	if s.token == "" && cache.Token != "" {
		s.token = cache.Token
	}

	if s.token == "" || !s.isTokenValid(s.token) {
		if !s.lastLoginAttempt.IsZero() && time.Since(s.lastLoginAttempt) < 2*time.Minute {
			if s.lastLoginError != nil {
				return fmt.Errorf("login on cooldown: %w", s.lastLoginError)
			}
			return fmt.Errorf("login on cooldown, please wait")
		}

		log.Println("[GoBizService] Token is invalid or missing, logging in...")
		if err := s.doLogin(); err != nil {
			return err
		}
	}

	if s.merchantID == "" && cache.MerchantID != "" {
		s.merchantID = cache.MerchantID
	}

	if s.merchantID == "" {
		log.Println("[GoBizService] Auto-detecting merchant ID...")
		merchants, err := s.getUserMerchants()
		if err != nil {
			return err
		}

		merchantID, err := extractMerchantID(merchants)
		if err != nil {
			return err
		}
		s.merchantID = merchantID

		// Update cache file
		cache = s.loadCache()
		cache.MerchantID = s.merchantID
		cache.Token = s.token
		s.saveCache(cache)
		log.Printf("[GoBizService] Auto-detected Merchant ID: %s\n", s.merchantID)
	}

	return nil
}

// getTransactionsAnalytics fetches transaction list via GoBiz analytics API.
func (s *GoBizService) getTransactionsAnalytics(days int) (map[string]interface{}, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	endTimeStr := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	startTimeStr := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02T15:04:05.000Z")

	urlStr := fmt.Sprintf("https://api.gojekapi.com/merchant-analytics/v2/merchants/transactions?from=0&size=50&statuses=SETTLEMENT,CAPTURE,REFUND,PARTIAL_REFUND&payment_types=QRIS,GOPAY,OFFLINE_CREDIT_CARD,OFFLINE_DEBIT_CARD,CREDIT_CARD&start_time=%s&end_time=%s&merchant_ids=%s",
		startTimeStr, endTimeStr, s.merchantID)

	headers := map[string]string{
		"accept":              "application/json, text/plain, */*",
		"accept-language":     "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7",
		"authentication-type": "go-id",
		"authorization":       "Bearer " + s.token,
		"content-type":        "application/json",
		"sec-ch-ua":           ``,
		"sec-fetch-dest":      "empty",
		"sec-fetch-mode":      "cors",
		"sec-fetch-site":      "cross-site",
	}

	respBytes, statusCode, err := execCurl("GET", urlStr, headers, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		log.Println("[GoBizService] Token expired during transactions fetch. Retrying login...")
		s.mu.Lock()
		s.token = "" // triggers doLogin during s.init()
		s.mu.Unlock()
		if err := s.init(); err != nil {
			return nil, err
		}
		// retry
		headers["authorization"] = "Bearer " + s.token
		respBytes, statusCode, err = execCurl("GET", urlStr, headers, nil)
		if err != nil {
			return nil, err
		}
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("transactions API returned status %d: %s", statusCode, string(respBytes))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// getTransactionsJournal fetches transaction list via GoBiz journal API.
func (s *GoBizService) getTransactionsJournal(days int) (map[string]interface{}, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	endTimeStr := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	startTimeStr := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02T15:04:05.000Z")

	requestBody := map[string]interface{}{
		"from": 0,
		"size": 50,
		"sort": map[string]interface{}{
			"time": map[string]interface{}{
				"order": "desc",
			},
		},
		"included_categories": map[string]interface{}{
			"incoming": []string{"transaction_share", "action"},
		},
		"query": []interface{}{
			map[string]interface{}{
				"op": "and",
				"clauses": []interface{}{
					map[string]interface{}{
						"op": "not",
						"clauses": []interface{}{
							map[string]interface{}{
								"op": "or",
								"clauses": []interface{}{
									map[string]interface{}{
										"field": "metadata.source",
										"op":    "in",
										"value": []string{"GOSAVE_ONLINE", "GoSave", "GODEALS_ONLINE"},
									},
									map[string]interface{}{
										"field": "metadata.gopay.source",
										"op":    "in",
										"value": []string{"GOSAVE_ONLINE", "GoSave", "GODEALS_ONLINE"},
									},
								},
							},
						},
					},
					map[string]interface{}{
						"field": "metadata.transaction.status",
						"op":    "in",
						"value": []string{"settlement", "capture", "refund", "partial_refund"},
					},
					map[string]interface{}{
						"op": "or",
						"clauses": []interface{}{
							map[string]interface{}{
								"op": "or",
								"clauses": []interface{}{
									map[string]interface{}{
										"field": "metadata.transaction.payment_type",
										"op":    "in",
										"value": []string{"qris", "gopay", "offline_credit_card", "offline_debit_card", "credit_card"},
									},
								},
							},
						},
					},
					map[string]interface{}{
						"field": "metadata.transaction.transaction_time",
						"op":    "gte",
						"value": startTimeStr,
					},
					map[string]interface{}{
						"field": "metadata.transaction.transaction_time",
						"op":    "lte",
						"value": endTimeStr,
					},
					map[string]interface{}{
						"field": "metadata.transaction.merchant_id",
						"op":    "equal",
						"value": s.merchantID,
					},
				},
			},
		},
	}

	headers := map[string]string{
		"accept":              "application/json, text/plain, */*, application/vnd.journal.v1+json",
		"accept-language":     "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7",
		"authentication-type": "go-id",
		"authorization":       "Bearer " + s.token,
		"content-type":        "application/json",
		"sec-ch-ua":           ``,
		"sec-fetch-dest":      "empty",
		"sec-fetch-mode":      "cors",
		"sec-fetch-site":      "cross-site",
	}

	respBytes, statusCode, err := execCurl("POST", "https://api.gobiz.co.id/journals/search", headers, requestBody)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		log.Println("[GoBizService] Token expired during journals fetch. Retrying login...")
		s.mu.Lock()
		s.token = ""
		s.mu.Unlock()
		if err := s.init(); err != nil {
			return nil, err
		}
		// retry
		headers["authorization"] = "Bearer " + s.token
		respBytes, statusCode, err = execCurl("POST", "https://api.gobiz.co.id/journals/search", headers, requestBody)
		if err != nil {
			return nil, err
		}
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("journals API returned status %d: %s", statusCode, string(respBytes))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// CheckGoBizPayment scans GoBiz transactions for a payment matching the expected price.
func CheckGoBizPayment(totalPrice int, orderCreatedAt time.Time) (bool, error) {
	// DISABLED: Automatic QRIS GoBiz verification disabled. Admin performs manual verification.
	return false, nil

	/*
	s := GetGoBizService()

	// Initialize first, failing early if login/auth fails
	if err := s.init(); err != nil {
		return false, fmt.Errorf("failed to initialize GoBiz session: %w", err)
	}

	// 1. Try checking the analytics API (recent 1 day)
	data, err := s.getTransactionsAnalytics(1)
	if err != nil {
		log.Printf("[GoBizService] Gagal memuat data transaksi dari Analytics: %v. Mencoba fallback ke Journal API...\n", err)
	} else {
		found := matchTransaction(data, "transactions", totalPrice, orderCreatedAt)
		if found {
			return true, nil
		}
	}

	// 2. Try checking the journal search API (fallback)
	journalData, err := s.getTransactionsJournal(1)
	if err != nil {
		log.Printf("[GoBizService] Gagal memuat data transaksi dari Journal: %v\n", err)
		return false, err
	}

	found := matchJournalTransaction(journalData, totalPrice, orderCreatedAt)
	return found, nil
	*/
}

func matchTransaction(data map[string]interface{}, key string, totalPrice int, orderCreatedAt time.Time) bool {
	transactionsRaw, ok := data[key]
	if !ok || transactionsRaw == nil {
		return false
	}

	transactions, ok := transactionsRaw.([]interface{})
	if !ok {
		return false
	}

	expectedAmountCents := totalPrice * 100

	for _, txRaw := range transactions {
		tx, ok := txRaw.(map[string]interface{})
		if !ok {
			continue
		}
		status, _ := tx["transaction_status"].(string)
		if status == "" {
			status, _ = tx["status"].(string)
		}
		if status != "SETTLEMENT" && status != "CAPTURE" {
			continue
		}

		paymentType, _ := tx["payment_type"].(string)
		if paymentType != "QRIS" && paymentType != "GOPAY" {
			continue
		}

		var grossAmount float64
		switch v := tx["gross_amount"].(type) {
		case float64:
			grossAmount = v
		case int:
			grossAmount = float64(v)
		case int64:
			grossAmount = float64(v)
		}

		if int(grossAmount) != expectedAmountCents {
			continue
		}

		timeStr, _ := tx["transaction_time"].(string)
		if timeStr == "" {
			continue
		}

		txTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			txTime, err = time.Parse("2006-01-02T15:04:05.000Z", timeStr)
			if err != nil {
				continue
			}
		}

		// Buffer of 10 minutes to allow clock differences/minor latencies
		if txTime.After(orderCreatedAt.Add(-10 * time.Minute)) {
			txID, _ := tx["id"].(string)
			if txID == "" {
				txID, _ = tx["transaction_id"].(string)
			}
			log.Printf("[GoBizService] Cocok (Analytics): ID=%v, Nominal=%d, Waktu=%s\n",
				txID, int(grossAmount)/100, timeStr)
			return true
		}
	}

	return false
}

func matchJournalTransaction(data map[string]interface{}, totalPrice int, orderCreatedAt time.Time) bool {
	journalDataRaw, ok := data["data"]
	if !ok || journalDataRaw == nil {
		return false
	}

	journalItems, ok := journalDataRaw.([]interface{})
	if !ok {
		return false
	}

	expectedAmountCents := totalPrice * 100

	for _, itemRaw := range journalItems {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			continue
		}

		metadata, _ := item["metadata"].(map[string]interface{})
		if metadata == nil {
			continue
		}

		tx, _ := metadata["transaction"].(map[string]interface{})
		if tx == nil {
			continue
		}
		status, _ := tx["transaction_status"].(string)
		if status == "" {
			status, _ = tx["status"].(string)
		}
		statusLower := strings.ToLower(status)
		if statusLower != "settlement" && statusLower != "capture" {
			continue
		}

		paymentType, _ := tx["payment_type"].(string)
		paymentTypeLower := strings.ToLower(paymentType)
		if paymentTypeLower != "qris" && paymentTypeLower != "gopay" {
			continue
		}

		var grossAmount float64
		switch v := tx["gross_amount"].(type) {
		case float64:
			grossAmount = v
		case int:
			grossAmount = float64(v)
		case int64:
			grossAmount = float64(v)
		}

		if int(grossAmount) != expectedAmountCents {
			continue
		}

		timeStr, _ := tx["transaction_time"].(string)
		if timeStr == "" {
			continue
		}

		txTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			txTime, err = time.Parse("2006-01-02T15:04:05.000Z", timeStr)
			if err != nil {
				continue
			}
		}

		if txTime.After(orderCreatedAt.Add(-10 * time.Minute)) {
			txID, _ := tx["id"].(string)
			if txID == "" {
				txID, _ = tx["transaction_id"].(string)
			}
			log.Printf("[GoBizService] Cocok (Journal): ID=%v, Nominal=%d, Waktu=%s\n",
				txID, int(grossAmount)/100, timeStr)
			return true
		}
	}

	return false
}
