package service

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

func (s *OrderService) SendInvoiceEmail(order *model.Order, customStatus ...string) error {
	user, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	go func() {
		domain := "gmail.com"
		parts := strings.Split(smtpUsername, "@")
		if len(parts) > 1 {
			domain = parts[1]
		}

		statusText := "Perlu Dibayar"
		if len(customStatus) > 0 && customStatus[0] != "" {
			statusText = customStatus[0]
		} else {
			switch strings.ToLower(order.Status) {
			case "cancelled":
				statusText = "Dibatalkan"
			case "waiting_confirmation":
				statusText = "Menunggu Konfirmasi"
			case "confirmed", "paid", "completed":
				statusText = "Pembayaran Berhasil"
			default:
				statusText = "Perlu Dibayar"
			}
		}

		var statusTitle, statusDesc, statusBadge, subjectSuffix, paymentInstructions string

		switch statusText {
		case "Dibatalkan":
			statusTitle = "Pesanan Dibatalkan"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, pesanan Anda dengan nomor invoice <strong>%s</strong> telah dibatalkan.", user.Name, order.Invoice)
			statusBadge = `<span style="display: inline-block; background-color: #fee2e2; color: #991b1b; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Dibatalkan</span>`
			subjectSuffix = "Dibatalkan"
			paymentInstructions = ""
		case "Menunggu Konfirmasi":
			statusTitle = "Menunggu Konfirmasi Pembayaran"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, terima kasih telah mengonfirmasi pembayaran Anda. Bukti transaksi Anda sedang diverifikasi oleh admin kami.", user.Name)
			statusBadge = `<span style="display: inline-block; background-color: #e0e7ff; color: #3730a3; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Menunggu Konfirmasi</span>`
			subjectSuffix = "Menunggu Konfirmasi"
			paymentInstructions = ""
		case "Lunas", "lunas":
			statusTitle = "Pembayaran Lunas"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, terima kasih. Pembayaran pesanan Anda telah diverifikasi dan dinyatakan <strong>Lunas</strong>. Pesanan Anda akan segera diproses.", user.Name)
			statusBadge = `<span style="display: inline-block; background-color: #d1fae5; color: #065f46; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Lunas</span>`
			subjectSuffix = "Lunas"
			paymentInstructions = ""
		case "processed", "Diproses", "diproses":
			statusTitle = "Pesanan Sedang Diproses"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, pesanan Anda dengan nomor invoice <strong>%s</strong> telah dikonfirmasi dan saat ini sedang dalam tahap pengerjaan/pengemasan.", user.Name, order.Invoice)
			statusBadge = `<span style="display: inline-block; background-color: #e0f2fe; color: #0369a1; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Diproses</span>`
			subjectSuffix = "Diproses"
			paymentInstructions = ""
		case "shipped", "Dikirim", "dikirim":
			statusTitle = "Pesanan Sedang Dikirim"
			shippingInfo := "kurir ekspedisi"
			if order.ShippingName != "" && order.ShippingName != "-" {
				shippingInfo = order.ShippingName
				if order.ShippingService != "" && order.ShippingService != "-" {
					shippingInfo += fmt.Sprintf(" (%s)", order.ShippingService)
				}
			}
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, pesanan Anda dengan nomor invoice <strong>%s</strong> telah dikirimkan menggunakan <strong>%s</strong>.", user.Name, order.Invoice, shippingInfo)
			statusBadge = `<span style="display: inline-block; background-color: #f3e8ff; color: #7e22ce; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Dikirim</span>`
			subjectSuffix = "Dikirim"
			paymentInstructions = ""
		case "delivered", "Terkirim", "terkirim":
			statusTitle = "Pesanan Telah Sampai"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, pesanan Anda dengan nomor invoice <strong>%s</strong> telah berhasil sampai di alamat tujuan.", user.Name, order.Invoice)
			statusBadge = `<span style="display: inline-block; background-color: #dcfce7; color: #15803d; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Terkirim</span>`
			subjectSuffix = "Terkirim"
			paymentInstructions = ""
		case "completed", "Selesai", "selesai":
			statusTitle = "Pesanan Selesai"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, pesanan Anda dengan nomor invoice <strong>%s</strong> telah dinyatakan selesai. Terima kasih telah mempercayai Digital Papyrus!", user.Name, order.Invoice)
			statusBadge = `<span style="display: inline-block; background-color: #ccfbf1; color: #0f766e; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Selesai</span>`
			subjectSuffix = "Selesai"
			paymentInstructions = ""
		case "Pembayaran Berhasil", "Berhasil":
			statusTitle = "Pembayaran Berhasil"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, terima kasih atas pesanan Anda. Kami telah menerima pembayaran Anda dan pesanan Anda sedang diproses.", user.Name)
			statusBadge = `<span style="display: inline-block; background-color: #d1fae5; color: #065f46; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Pembayaran Berhasil</span>`
			subjectSuffix = "Pembayaran Berhasil"
			paymentInstructions = ""
		default: // "Perlu Dibayar"
			statusTitle = "Pesanan Perlu Dibayar"
			statusDesc = fmt.Sprintf("Halo <strong>%s</strong>, terima kasih atas pesanan Anda. Silakan segera lakukan pembayaran transfer bank sesuai nominal yang tertera dalam waktu maksimal 24 jam agar pesanan Anda dapat segera kami proses.", user.Name)
			statusBadge = `<span style="display: inline-block; background-color: #fef3c7; color: #92400e; padding: 4px 12px; border-radius: 9999px; font-weight: 700; font-size: 13px;">Perlu Dibayar</span>`
			subjectSuffix = "Perlu Dibayar"
			paymentInstructions = fmt.Sprintf(`
			<div style="background-color: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 12px; padding: 18px 20px; margin-bottom: 24px;">
			  <p style="margin: 0 0 10px 0; color: #166534; font-weight: 700; font-size: 14px;">Petunjuk Pembayaran Transfer Bank:</p>
			  <table style="width: 100%%; font-size: 13px; color: #166534;">
				<tr><td style="padding: 4px 0; width: 140px; font-weight: 600;">Bank Tujuan:</td><td style="font-weight: 700; color: #0f172a;">Bank BCA</td></tr>
				<tr><td style="padding: 4px 0; font-weight: 600;">Nomor Rekening:</td><td style="font-weight: 800; font-size: 16px; color: #14383f; letter-spacing: 0.5px;">8735091234</td></tr>
				<tr><td style="padding: 4px 0; font-weight: 600;">Atas Nama:</td><td style="font-weight: 700; color: #0f172a;">PT Digital Papyrus Indonesia</td></tr>
				<tr><td style="padding: 4px 0; font-weight: 600;">Jumlah Transfer:</td><td style="font-weight: 800; font-size: 16px; color: #14383f;">Rp %s</td></tr>
			  </table>
			  <p style="margin: 10px 0 0 0; font-size: 11px; color: #15803d; line-height: 1.4;">*Mohon transfer tepat sesuai nominal di atas (termasuk kode unik) untuk mempercepat verifikasi admin, kemudian klik 'Konfirmasi Pembayaran' di dashboard atau hubungi WhatsApp CS.</p>
			</div>
			`, formatRupiah(order.TotalPrice))
		}

		var itemsHTMLBuilder strings.Builder
		for _, item := range order.Items {
			itemType := "Jasa"
			if item.BookID != "" {
				itemType = "Buku Fisik"
				if strings.Contains(strings.ToLower(item.Format), "ebook") || strings.Contains(strings.ToLower(item.Format), "e-book") {
					itemType = "E-Book"
				}
			}
			itemsHTMLBuilder.WriteString(fmt.Sprintf(`
			<tr>
			  <td>
				<span class="item-title">%s</span>
				<span class="item-type">%s</span>
			  </td>
			  <td align="center">%d</td>
			  <td align="right">Rp %s</td>
			</tr>
			`, item.ItemName, itemType, item.Qty, formatRupiah(item.TotalPrice)))
		}

		// Resolve region names using our UserRepository helper
		provinceName := s.userRepo.GetRegionName("provinces", user.Province)
		cityName := s.userRepo.GetRegionName("regencies", user.City)
		districtName := s.userRepo.GetRegionName("districts", user.Regency)
		villageName := s.userRepo.GetRegionName("villages", user.Village)

		body := InvoiceEmailTemplate
		body = strings.Replace(body, "STATUS_TITLE_PLACEHOLDER", statusTitle, 1)
		body = strings.Replace(body, "STATUS_DESC_PLACEHOLDER", statusDesc, 1)
		body = strings.Replace(body, "STATUS_BADGE_PLACEHOLDER", statusBadge, 1)
		body = strings.Replace(body, "PAYMENT_INSTRUCTIONS_PLACEHOLDER", paymentInstructions, 1)
		body = strings.Replace(body, "CUSTOMER_NAME_PLACEHOLDER", user.Name, -1)
		body = strings.Replace(body, "INVOICE_PLACEHOLDER", order.Invoice, -1)
		loc := time.FixedZone("WIB", 7*60*60)
		formattedDate := order.CreatedAt.In(loc).Format("02 Jan 2006 15:04") + " WIB"
		body = strings.Replace(body, "DATE_PLACEHOLDER", formattedDate, 1)
		body = strings.Replace(body, "PAYMENT_TYPE_PLACEHOLDER", strings.ToUpper(order.PaymentType), 1)
		body = strings.Replace(body, "ADDRESS_PLACEHOLDER", user.Address, 1)
		body = strings.Replace(body, "VILLAGE_PLACEHOLDER", villageName, 1)
		body = strings.Replace(body, "DISTRICT_PLACEHOLDER", districtName, 1)
		body = strings.Replace(body, "CITY_PLACEHOLDER", cityName, 1)
		body = strings.Replace(body, "PROVINCE_PLACEHOLDER", provinceName, 1)
		body = strings.Replace(body, "ZIP_PLACEHOLDER", user.ZipCode, 1)
		body = strings.Replace(body, "ITEMS_PLACEHOLDER", itemsHTMLBuilder.String(), 1)

		itemsSubtotal := 0
		for _, item := range order.Items {
			itemsSubtotal += item.TotalPrice
		}
		uniqueCode := order.TotalPrice - itemsSubtotal - order.ShippingPrice - order.ServiceFee - order.Tax + order.Discount
		if uniqueCode < 0 {
			uniqueCode = 0
		}

		var summaryBuilder strings.Builder
		
		// 1. Subtotal
		summaryBuilder.WriteString(fmt.Sprintf(`
		<tr class="summary-row">
		  <td width="70%%" style="padding: 8px 16px; color: #516265; text-align: right; font-weight: 600;">Subtotal</td>
		  <td width="30%%" style="padding: 8px 16px; color: #516265; text-align: right;">Rp %s</td>
		</tr>
		`, formatRupiah(itemsSubtotal)))
		
		// 2. Shipping
		shippingInfo := fmt.Sprintf("%s (%s)", order.ShippingName, order.ShippingService)
		if order.ShippingName == "" || order.ShippingName == "-" {
			shippingInfo = "Gratis Ongkir / Kurir"
		}
		summaryBuilder.WriteString(fmt.Sprintf(`
		<tr class="summary-row">
		  <td style="padding: 8px 16px; color: #516265; text-align: right; font-weight: 600;">Ongkos Kirim (%s)</td>
		  <td style="padding: 8px 16px; color: #516265; text-align: right;">Rp %s</td>
		</tr>
		`, shippingInfo, formatRupiah(order.ShippingPrice)))
		
		// 3. Service Fee
		if order.ServiceFee > 0 {
			summaryBuilder.WriteString(fmt.Sprintf(`
			<tr class="summary-row">
			  <td style="padding: 8px 16px; color: #516265; text-align: right; font-weight: 600;">Biaya Layanan</td>
			  <td style="padding: 8px 16px; color: #516265; text-align: right;">Rp %s</td>
			</tr>
			`, formatRupiah(order.ServiceFee)))
		}
		
		// 4. Tax
		if order.Tax > 0 {
			summaryBuilder.WriteString(fmt.Sprintf(`
			<tr class="summary-row">
			  <td style="padding: 8px 16px; color: #516265; text-align: right; font-weight: 600;">Pajak</td>
			  <td style="padding: 8px 16px; color: #516265; text-align: right;">Rp %s</td>
			</tr>
			`, formatRupiah(order.Tax)))
		}
		
		// 5. Discount
		if order.Discount > 0 {
			summaryBuilder.WriteString(fmt.Sprintf(`
			<tr class="summary-row" style="color: #dc2626;">
			  <td style="padding: 8px 16px; text-align: right; font-weight: 600;">Diskon</td>
			  <td style="padding: 8px 16px; text-align: right;">-Rp %s</td>
			</tr>
			`, formatRupiah(order.Discount)))
		}
		
		// 6. Unique Code
		if uniqueCode > 0 {
			summaryBuilder.WriteString(fmt.Sprintf(`
			<tr class="summary-row">
			  <td style="padding: 8px 16px; color: #b45309; text-align: right; font-weight: 600;">Kode Unik</td>
			  <td style="padding: 8px 16px; color: #b45309; text-align: right; font-weight: 700;">Rp %s</td>
			</tr>
			`, formatRupiah(uniqueCode)))
		}
		
		// 7. Grand Total
		summaryBuilder.WriteString(fmt.Sprintf(`
		<tr class="summary-total">
		  <td style="padding: 16px; font-size: 18px; font-weight: 800; color: #14383f; border-top: 2px solid #eef4f5; background-color: #f9fbfc; text-align: right;">Total Pembayaran</td>
		  <td style="padding: 16px; font-size: 18px; font-weight: 800; color: #14383f; border-top: 2px solid #eef4f5; background-color: #f9fbfc; text-align: right;">Rp %s</td>
		</tr>
		`, formatRupiah(order.TotalPrice)))

		body = strings.Replace(body, "SUMMARY_ROWS_PLACEHOLDER", summaryBuilder.String(), 1)

		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("From: Digital Papyrus <%s>\r\n", smtpUsername))
		msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", user.Email))
		msgBuilder.WriteString("Cc: digitalpapyrus15@gmail.com\r\n")
		msgBuilder.WriteString(fmt.Sprintf("Subject: Invoice Pesanan %s - %s\r\n", order.Invoice, subjectSuffix))
		msgBuilder.WriteString(fmt.Sprintf("Message-ID: <%d@%s>\r\n", time.Now().UnixNano(), domain))
		msgBuilder.WriteString("MIME-version: 1.0;\r\n")
		msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
		msgBuilder.WriteString(body)

		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		err := smtp.SendMail(addr, auth, smtpUsername, []string{user.Email, "digitalpapyrus15@gmail.com"}, []byte(msgBuilder.String()))
		if err != nil {
			log.Printf("[SMTP ERROR] Failed to send invoice to %s and CC: %v", user.Email, err)
		} else {
			log.Printf("[SMTP SUCCESS] Sent invoice (%s) to %s (Invoice: %s)", statusText, user.Email, order.Invoice)
		}
	}()

	return nil
}

func formatRupiah(amount int) string {
	s := strconv.Itoa(amount)
	var res []byte
	for i := len(s) - 1; i >= 0; i-- {
		res = append([]byte{s[i]}, res...)
		if (len(s)-i)%3 == 0 && i != 0 {
			res = append([]byte{'.'}, res...)
		}
	}
	return string(res)
}
