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
