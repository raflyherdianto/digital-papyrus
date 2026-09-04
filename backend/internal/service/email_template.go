package service

// OTPEmailTemplate is the HTML template used for sending OTP email verifications.
const OTPEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verifikasi Email Anda — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 580px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 22px;
      font-weight: 700;
      color: #14383f;
      margin: 0 0 16px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 24px 0;
      text-align: center;
    }
    
    /* OTP Code Box */
    .otp-container {
      background: linear-gradient(135deg, #f0f7f8 0%, #e8f3f5 100%);
      border: 2px dashed #79a8af;
      border-radius: 12px;
      padding: 24px;
      margin: 32px 0;
      text-align: center;
    }
    .otp-code {
      font-family: 'Courier New', Courier, monospace;
      font-size: 38px;
      font-weight: 800;
      letter-spacing: 8px;
      color: #14383f;
      margin: 0;
      text-shadow: 0 1px 2px rgba(255, 255, 255, 0.9);
    }
    .otp-expires {
      font-size: 13px;
      color: #738a8e;
      font-weight: 500;
      margin-top: 10px;
      margin-bottom: 0;
    }
    
    /* Warning/Security block */
    .security-note {
      font-size: 13px;
      color: #8c9fa2;
      background-color: #fafcfc;
      border-left: 3px solid #79a8af;
      padding: 12px 16px;
      border-radius: 0 8px 8px 0;
      margin-top: 24px;
    }
    
    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Verifikasi Akun Baru Anda</h1>
            <p class="instructions">
              Terima kasih telah mendaftar di <strong>Digital Papyrus</strong>. Silakan gunakan kode verifikasi di bawah ini untuk menyelesaikan proses pendaftaran akun Anda:
            </p>
            
            <!-- OTP Box -->
            <div class="otp-container">
              <div class="otp-code">184920</div>
              <p class="otp-expires">Kode berlaku selama <strong>5 menit</strong></p>
            </div>
            
            <p class="instructions" style="margin-top: 24px;">
              Masukkan kode ini di halaman pendaftaran untuk memverifikasi alamat email Anda.
            </p>
            
            <div class="security-note">
              <strong>Catatan Keamanan:</strong> Jangan bagikan kode verifikasi ini kepada siapa pun. Tim Digital Papyrus tidak pernah meminta kode sandi atau OTP Anda. Jika Anda tidak merasa melakukan pendaftaran ini, abaikan email ini.
            </div>
          </div>
          
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// ResetPasswordEmailTemplate is the HTML template used for sending password reset links.
const ResetPasswordEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Atur Ulang Kata Sandi — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 580px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 22px;
      font-weight: 700;
      color: #14383f;
      margin: 0 0 16px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 24px 0;
      text-align: center;
    }
    
    /* Button */
    .button-container {
      text-align: center;
      margin: 32px 0;
    }
    .button {
      background-color: #14383f;
      color: #ffffff !important;
      padding: 14px 30px;
      text-decoration: none;
      font-weight: 700;
      border-radius: 8px;
      font-size: 14px;
      display: inline-block;
      box-shadow: 0 4px 12px rgba(20, 56, 63, 0.15);
    }
    
    /* Warning/Security block */
    .security-note {
      font-size: 13px;
      color: #8c9fa2;
      background-color: #fafcfc;
      border-left: 3px solid #79a8af;
      padding: 12px 16px;
      border-radius: 0 8px 8px 0;
      margin-top: 24px;
    }
    
    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Atur Ulang Kata Sandi</h1>
            <p class="instructions">
              Anda menerima email ini karena kami menerima permintaan untuk mengatur ulang kata sandi akun <strong>Digital Papyrus</strong> Anda. Silakan klik tombol di bawah ini untuk mengatur kata sandi baru Anda:
            </p>
            
            <!-- Button -->
            <div class="button-container">
              <a href="RESET_LINK_PLACEHOLDER" class="button" target="_blank">Atur Ulang Kata Sandi</a>
            </div>
            
            <p class="instructions" style="margin-top: 24px;">
              Tautan ini hanya berlaku selama <strong>1 jam</strong>.
            </p>
            
            <div class="security-note">
              <strong>Catatan Keamanan:</strong> Jika Anda tidak meminta pengaturan ulang kata sandi ini, Anda dapat mengabaikan email ini dengan aman. Kata sandi Anda akan tetap aman.
            </div>
          </div>
          
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// PasswordChangedEmailTemplate is the HTML template used for sending password changed confirmation.
const PasswordChangedEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Kata Sandi Berhasil Diperbarui — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 580px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 22px;
      font-weight: 700;
      color: #14383f;
      margin: 0 0 16px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 24px 0;
      text-align: center;
    }
    .details-box {
      background-color: #f8fafb;
      border: 1px solid #eef2f3;
      border-radius: 12px;
      padding: 20px;
      margin: 24px 0;
    }
    .details-item {
      font-size: 14px;
      margin: 8px 0;
      color: #4a5557;
    }
    .details-label {
      font-weight: 600;
      color: #14383f;
    }
    .security-note {
      font-size: 13px;
      color: #8c9fa2;
      background-color: #fafcfc;
      border-left: 3px solid #d9534f;
      padding: 12px 16px;
      border-radius: 0 8px 8px 0;
      margin-top: 24px;
    }
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Kata Sandi Berhasil Diperbarui</h1>
            <p class="instructions">
              Halo, pemberitahuan ini dikirim untuk mengonfirmasi bahwa kata sandi akun <strong>Digital Papyrus</strong> Anda telah berhasil diubah.
            </p>
            
            <div class="details-box">
              <div class="details-item"><span class="details-label">Waktu Pembaruan:</span> DATE_PLACEHOLDER WIB</div>
              <div class="details-item"><span class="details-label">Alamat Email:</span> EMAIL_PLACEHOLDER</div>
              <div class="details-item"><span class="details-label">Status:</span> Berhasil Diperbarui</div>
            </div>
            
            <div class="security-note">
              <strong>PENTING:</strong> Jika Anda tidak melakukan perubahan ini, akun Anda mungkin dalam bahaya. Segera hubungi Customer Service kami melalui WhatsApp di <a href="https://wa.me/6285196206398" style="color: #d9534f; font-weight: 600;">0851 9620 6398</a> atau ubah kembali kata sandi Anda segera melalui fitur "Forgot Password".
            </div>
          </div>
          
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// InvoiceEmailTemplate is the HTML template used for sending transaction invoices.
const InvoiceEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Invoice Pesanan — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 600px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 24px;
      font-weight: 700;
      color: #14383f;
      margin: 0 0 8px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 30px 0;
      text-align: center;
    }

    /* Details Section */
    .details-box {
      background-color: #f9fbfc;
      border: 1px solid #eef4f5;
      border-radius: 12px;
      padding: 24px;
      margin-bottom: 30px;
    }
    .details-label {
      color: #8c9fa2;
      font-weight: 600;
      text-transform: uppercase;
      font-size: 11px;
      letter-spacing: 0.05em;
    }
    .details-value {
      color: #14383f;
      font-weight: 700;
    }
    .details-address {
      color: #14383f;
      font-weight: 500;
      font-size: 13px;
      margin-top: 4px;
      line-height: 1.4;
    }

    /* Items Table */
    .items-table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 30px;
    }
    .items-table th {
      text-align: left;
      padding: 12px 16px;
      background-color: #f0f7f8;
      color: #516265;
      font-size: 12px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      border-bottom: 2px solid #eef4f5;
    }
    .items-table td {
      padding: 16px;
      border-bottom: 1px solid #eef4f5;
      font-size: 14px;
      color: #2c3539;
    }
    .items-table .item-title {
      font-weight: 700;
      color: #14383f;
      display: block;
      margin-bottom: 4px;
    }
    .items-table .item-type {
      font-size: 11px;
      color: #79a8af;
      text-transform: uppercase;
      font-weight: 600;
      letter-spacing: 0.05em;
    }

    /* Summary Table */
    .summary-table {
      width: 100%;
      margin-bottom: 30px;
    }
    .summary-row {
      font-size: 14px;
    }
    .summary-row td {
      padding: 8px 16px;
      color: #516265;
      text-align: right;
    }
    .summary-row td:first-child {
      font-weight: 600;
    }
    .summary-total td {
      padding: 16px;
      font-size: 18px;
      font-weight: 800;
      color: #14383f;
      border-top: 2px solid #eef4f5;
      background-color: #f9fbfc;
    }

    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
    .details-table {
      width: 100%;
    }
    .details-col {
      vertical-align: top;
    }
    
    @media only screen and (max-width: 480px) {
      .email-container {
        margin: 10px auto !important;
        border-radius: 8px !important;
      }
      .content {
        padding: 24px 20px 20px 20px !important;
      }
      .details-box {
        padding: 16px !important;
      }
      .items-table th, .items-table td {
        padding: 12px 10px !important;
      }
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">STATUS_TITLE_PLACEHOLDER</h1>
            <p class="instructions">
              STATUS_DESC_PLACEHOLDER
            </p>

            PAYMENT_INSTRUCTIONS_PLACEHOLDER

            <div class="details-box">
              <table class="details-table" width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td class="details-col" valign="top" style="padding-bottom: 12px; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Nomor Invoice</div>
                    <div class="details-value" style="text-align: left; font-family: monospace; font-size: 15px; margin-top: 2px;">INVOICE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Status Pesanan</div>
                    <div class="details-value" style="text-align: left; margin-top: 4px;">STATUS_BADGE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Tanggal</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">DATE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Metode Pembayaran</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">PAYMENT_TYPE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding-top: 12px; text-align: left;">
                    <div class="details-label" style="text-align: left;">Alamat Pengiriman</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px; margin-bottom: 4px;">CUSTOMER_NAME_PLACEHOLDER</div>
                    <div class="details-address" style="text-align: left;">
                      ADDRESS_PLACEHOLDER<br>
                      VILLAGE_PLACEHOLDER, DISTRICT_PLACEHOLDER<br>
                      CITY_PLACEHOLDER, PROVINCE_PLACEHOLDER, ZIP_PLACEHOLDER
                    </div>
                  </td>
                </tr>
              </table>
            </div>
            
            <table class="items-table">
              <thead>
                <tr>
                  <th>Item</th>
                  <th style="text-align: center;">Qty</th>
                  <th style="text-align: right;">Total</th>
                </tr>
              </thead>
              <tbody>
                ITEMS_PLACEHOLDER
              </tbody>
            </table>

            <table class="summary-table">
              SUMMARY_ROWS_PLACEHOLDER
            </table>

            <p style="text-align: center; color: #8c9fa2; font-size: 13px; margin-top: 40px;">
              Anda dapat melacak status pesanan Anda melalui <a href="https://digitalpapyrus.web.id/customer-dashboard" style="color: #14383f; font-weight: 600; text-decoration: underline;">Dashboard Pelanggan</a>.
            </p>
          </div>
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// DraftSubmissionEmailTemplate is the HTML template used for draft submission notifications.
const DraftSubmissionEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Upload Draft Berhasil — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 600px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 24px;
      font-weight: 700;
      color: #14383f;
      margin: 0 0 8px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 30px 0;
      text-align: center;
    }

    /* Details Section */
    .details-box {
      background-color: #f9fbfc;
      border: 1px solid #eef4f5;
      border-radius: 12px;
      padding: 24px;
      margin-bottom: 30px;
    }
    .details-label {
      color: #8c9fa2;
      font-weight: 600;
      text-transform: uppercase;
      font-size: 11px;
      letter-spacing: 0.05em;
    }
    .details-value {
      color: #14383f;
      font-weight: 700;
    }
    .details-address {
      color: #14383f;
      font-weight: 500;
      font-size: 13px;
      margin-top: 4px;
      line-height: 1.4;
    }

    /* Items Table */
    .items-table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 30px;
    }
    .items-table th {
      text-align: left;
      padding: 12px 16px;
      background-color: #f0f7f8;
      color: #516265;
      font-size: 12px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      border-bottom: 2px solid #eef4f5;
    }
    .items-table td {
      padding: 16px;
      border-bottom: 1px solid #eef4f5;
      font-size: 14px;
      color: #2c3539;
    }
    .items-table .item-title {
      font-weight: 700;
      color: #14383f;
      display: block;
      margin-bottom: 4px;
    }
    .items-table .item-type {
      font-size: 11px;
      color: #79a8af;
      text-transform: uppercase;
      font-weight: 600;
      letter-spacing: 0.05em;
    }

    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
    .details-table {
      width: 100%;
    }
    .details-col {
      vertical-align: top;
    }
    
    @media only screen and (max-width: 480px) {
      .email-container {
        margin: 10px auto !important;
        border-radius: 8px !important;
      }
      .content {
        padding: 24px 20px 20px 20px !important;
      }
      .details-box {
        padding: 16px !important;
      }
      .items-table th, .items-table td {
        padding: 12px 10px !important;
      }
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Upload Draft Berhasil</h1>
            <p class="instructions">
              Halo <strong>CUSTOMER_NAME_PLACEHOLDER</strong>, terima kasih atas pengajuan draf buku Anda. Kami telah menerima berkas draf Anda dan tim admin kami akan segera meninjaunya. Mohon menunggu persetujuan admin.
            </p>
            
            <div class="details-box">
              <table class="details-table" width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td class="details-col" valign="top" style="padding-bottom: 12px; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Nomor Invoice</div>
                    <div class="details-value" style="text-align: left; font-family: monospace; font-size: 15px; margin-top: 2px;">INVOICE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Judul Buku</div>
                    <div class="details-value" style="text-align: left; font-size: 15px; margin-top: 2px;">BOOK_TITLE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Penulis</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_AUTHOR_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Format</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_FORMAT_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding-top: 12px; text-align: left;">
                    <div class="details-label" style="text-align: left;">Status Validasi</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px; color: #d97706;">Pending Validation</div>
                  </td>
                </tr>
              </table>
            </div>

            <p style="text-align: center; color: #8c9fa2; font-size: 13px; margin-top: 40px;">
              Anda dapat memantau proses validasi melalui menu "View Progress" pada <a href="https://digitalpapyrus.web.id/customer-publish" style="color: #14383f; font-weight: 600; text-decoration: underline;">Halaman Publish</a>.
            </p>
          </div>
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// ValidationApproveEmailTemplate is the HTML template used for approved draft notifications.
const ValidationApproveEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Draft Disetujui — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 600px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 24px;
      font-weight: 700;
      color: #16a34a;
      margin: 0 0 8px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 30px 0;
      text-align: center;
    }

    /* Details Section */
    .details-box {
      background-color: #f9fbfc;
      border: 1px solid #eef4f5;
      border-radius: 12px;
      padding: 24px;
      margin-bottom: 30px;
    }
    .details-label {
      color: #8c9fa2;
      font-weight: 600;
      text-transform: uppercase;
      font-size: 11px;
      letter-spacing: 0.05em;
    }
    .details-value {
      color: #14383f;
      font-weight: 700;
    }
    .details-table {
      width: 100%;
    }
    .details-col {
      vertical-align: top;
    }
    
    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
    
    @media only screen and (max-width: 480px) {
      .email-container {
        margin: 10px auto !important;
        border-radius: 8px !important;
      }
      .content {
        padding: 24px 20px 20px 20px !important;
      }
      .details-box {
        padding: 16px !important;
      }
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Draft Disetujui</h1>
            <p class="instructions">
              Halo <strong>CUSTOMER_NAME_PLACEHOLDER</strong>, selamat! Draf buku Anda telah disetujui oleh tim kami dan sekarang sedang diproses untuk publikasi.
            </p>
            
            <div class="details-box">
              <table class="details-table" width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td class="details-col" valign="top" style="padding-bottom: 12px; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Nomor Invoice</div>
                    <div class="details-value" style="text-align: left; font-family: monospace; font-size: 15px; margin-top: 2px;">INVOICE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Judul Buku</div>
                    <div class="details-value" style="text-align: left; font-size: 15px; margin-top: 2px;">BOOK_TITLE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Penulis</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_AUTHOR_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Format</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_FORMAT_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding-top: 12px; text-align: left;">
                    <div class="details-label" style="text-align: left;">Status Validasi</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px; color: #16a34a;">Approved</div>
                  </td>
                </tr>
              </table>
            </div>

            <p style="text-align: center; color: #8c9fa2; font-size: 13px; margin-top: 40px;">
              Anda dapat memantau proses validasi melalui menu "View Progress" pada <a href="https://digitalpapyrus.web.id/customer-publish" style="color: #14383f; font-weight: 600; text-decoration: underline;">Halaman Publish</a>.
            </p>
          </div>
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`

// ValidationRejectEmailTemplate is the HTML template used for rejected draft notifications.
const ValidationRejectEmailTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Draft Ditolak — Digital Papyrus</title>
  <style>
    /* Reset and basics */
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f4f7f8;
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    table {
      border-spacing: 0;
      border-collapse: collapse;
      width: 100%;
    }
    td {
      font-family: inherit;
    }
    img {
      border: 0;
      outline: none;
      text-decoration: none;
      max-width: 100%;
    }
    
    /* Layout styling */
    .email-container {
      max-width: 600px;
      margin: 40px auto;
      background-color: #ffffff;
      border-radius: 16px;
      overflow: hidden;
      box-shadow: 0 4px 24px rgba(20, 56, 63, 0.06);
      border: 1px solid #e2ecee;
    }
    
    /* Brand Header */
    .header {
      background-color: #ffffff; 
      padding: 36px 32px;
      text-align: center;
      border-bottom: 2px solid #eef4f5;
    }
    .header-logo-img {
      height: 80px;
      width: auto;
      display: block;
      margin: 0 auto 12px auto;
      max-width: 100%;
    }
    .header-logo {
      color: #14383f;
      font-size: 22px;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      margin: 0;
    }
    .header-tagline {
      color: #79a8af;
      font-size: 11px;
      margin: 4px 0 0 0;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      font-weight: 600;
    }
    
    /* Body Content */
    .content {
      padding: 40px 40px 30px 40px;
      color: #2c3539;
      line-height: 1.6;
    }
    .welcome-text {
      font-size: 24px;
      font-weight: 700;
      color: #dc2626;
      margin: 0 0 8px 0;
      text-align: center;
    }
    .instructions {
      font-size: 15px;
      color: #516265;
      margin: 0 0 30px 0;
      text-align: center;
    }

    /* Details Section */
    .details-box {
      background-color: #f9fbfc;
      border: 1px solid #eef4f5;
      border-radius: 12px;
      padding: 24px;
      margin-bottom: 30px;
    }
    .details-label {
      color: #8c9fa2;
      font-weight: 600;
      text-transform: uppercase;
      font-size: 11px;
      letter-spacing: 0.05em;
    }
    .details-value {
      color: #14383f;
      font-weight: 700;
    }
    .details-table {
      width: 100%;
    }
    .details-col {
      vertical-align: top;
    }
    
    /* Footer */
    .footer {
      background-color: #fafcfc;
      border-top: 1px solid #eef4f5;
      padding: 32px 40px;
      text-align: center;
      font-size: 12px;
      color: #8c9fa2;
      line-height: 1.5;
    }
    .footer a {
      color: #14383f;
      text-decoration: none;
      font-weight: 600;
    }
    .social-links {
      margin-top: 16px;
    }
    .social-links a {
      margin: 0 8px;
      color: #79a8af;
    }
    
    @media only screen and (max-width: 480px) {
      .email-container {
        margin: 10px auto !important;
        border-radius: 8px !important;
      }
      .content {
        padding: 24px 20px 20px 20px !important;
      }
      .details-box {
        padding: 16px !important;
      }
    }
  </style>
</head>
<body>
  <table role="presentation" width="100%">
    <tr>
      <td align="center" style="padding: 20px 10px;">
        <div class="email-container">
          <!-- Header -->
          <div class="header">
            <img src="https://digitalpapyrus.web.id/logo.png" alt="Digital Papyrus Logo" class="header-logo-img" />
            <div class="header-logo">Digital Papyrus</div>
            <div class="header-tagline">Warisan Karya Bertemu Era Digital</div>
          </div>
          
          <!-- Content -->
          <div class="content">
            <h1 class="welcome-text">Draft Ditolak</h1>
            <p class="instructions">
              Halo <strong>CUSTOMER_NAME_PLACEHOLDER</strong>, mohon maaf, draf buku Anda saat ini tidak dapat disetujui oleh tim kami. Silakan periksa catatan dari tim kami di bawah ini.
            </p>
            
            <div class="details-box">
              <table class="details-table" width="100%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td class="details-col" valign="top" style="padding-bottom: 12px; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Nomor Invoice</div>
                    <div class="details-value" style="text-align: left; font-family: monospace; font-size: 15px; margin-top: 2px;">INVOICE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Judul Buku</div>
                    <div class="details-value" style="text-align: left; font-size: 15px; margin-top: 2px;">BOOK_TITLE_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Penulis</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_AUTHOR_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Format</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px;">BOOK_FORMAT_PLACEHOLDER</div>
                  </td>
                </tr>
                <tr>
                  <td class="details-col" valign="top" style="padding: 12px 0; border-bottom: 1px solid #eef4f5; text-align: left;">
                    <div class="details-label" style="text-align: left;">Status Validasi</div>
                    <div class="details-value" style="text-align: left; margin-top: 2px; color: #dc2626;">Rejected</div>
                  </td>
                </tr>
              </table>
              <div class="notes-box" style="background-color: #fff1f2; border-left: 4px solid #f43f5e; padding: 16px; margin-top: 20px; font-size: 14px; color: #9f1239; text-align: left; border-radius: 8px;">
                <strong>Catatan dari Admin:</strong><br/>
                ADMIN_NOTES_PLACEHOLDER
              </div>
            </div>

            <p style="text-align: center; color: #8c9fa2; font-size: 13px; margin-top: 40px;">
              Anda dapat memantau proses validasi melalui menu "View Progress" pada <a href="https://digitalpapyrus.web.id/customer-publish" style="color: #14383f; font-weight: 600; text-decoration: underline;">Halaman Publish</a>.
            </p>
          </div>
          <!-- Footer -->
          <div class="footer">
            <p style="margin: 0 0 10px 0;">
              &copy; 2026 <strong>Digital Papyrus</strong>. All rights reserved.
            </p>
            <p style="margin: 0;">
              Butuh bantuan? Hubungi tim CS kami di 
              <a href="https://wa.me/6285196206398" target="_blank">WhatsApp (0851 9620 6398)</a> atau email ke 
              <a href="mailto:digitalpapyrus15@gmail.com">digitalpapyrus15@gmail.com</a>.
            </p>
            <div class="social-links">
              <a href="https://digitalpapyrus.web.id/" target="_blank">Website</a> &bull;
              <a href="https://digitalpapyrus.web.id/katalog" target="_blank">Katalog Buku</a>
            </div>
          </div>
        </div>
      </td>
    </tr>
  </table>
</body>
</html>`
