// Package email is a tiny SMTP sender for meet's operator notifications.
//
// Added by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
// Timeful sent mail via Listmonk + Google Cloud Tasks, both disabled in this
// fork. This replaces them with plain SMTP (e.g. a Gmail app password) so the
// operator can be notified when someone responds and invitees can be emailed a
// poll link. Everything is a no-op unless SMTP_* is configured, mirroring how
// the listmonk service degrades gracefully.
package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"schej.it/server/logger"
)

// Enabled reports whether SMTP delivery is configured.
func Enabled() bool {
	return os.Getenv("SMTP_HOST") != "" &&
		os.Getenv("SMTP_USER") != "" &&
		os.Getenv("SMTP_PASS") != ""
}

// OperatorEmail is the default recipient for owner notifications — used for
// Mensa-created events, which have no owner user to look up.
func OperatorEmail() string {
	return strings.TrimSpace(os.Getenv("OPERATOR_EMAIL"))
}

// Send delivers a single HTML email. It is a no-op (with no error) when SMTP is
// not configured or the recipient is empty, and it never panics — send failures
// are logged, since email is best-effort and must not break a response.
func Send(to, subject, htmlBody string) {
	if !Enabled() || strings.TrimSpace(to) == "" {
		return
	}

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `text/html; charset="UTF-8"`,
	}
	var b strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&b, "%s: %s\r\n", k, v)
	}
	b.WriteString("\r\n")
	b.WriteString(htmlBody)

	// smtp.SendMail upgrades to STARTTLS automatically when the server offers it
	// (Gmail's smtp.gmail.com:587 does), so an app password works out of the box.
	auth := smtp.PlainAuth("", user, pass, host)
	addr := fmt.Sprintf("%s:%s", host, port)
	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(b.String())); err != nil {
		logger.StdErr.Printf("email: send to %s failed: %v", to, err)
	}
}

// SimpleBody renders a minimal dark-themed HTML email: a greeting, a line of
// body text (may contain inline HTML), and a call-to-action button linking to
// url.
func SimpleBody(greetingName, bodyHTML, url, ctaLabel string) string {
	greeting := "Hi"
	if strings.TrimSpace(greetingName) != "" {
		greeting = "Hi " + greetingName
	}
	return fmt.Sprintf(`<div style="font-family:'Spectral',Georgia,serif;background:#0a0a0a;color:#f4f4f4;padding:28px;border-radius:12px;max-width:520px;margin:auto">
  <div style="font-size:22px;font-weight:700;letter-spacing:1px;margin-bottom:16px">meet with jdh</div>
  <p style="font-size:15px;line-height:1.6;margin:0 0 8px">%s,</p>
  <p style="font-size:15px;line-height:1.6;margin:0 0 20px">%s</p>
  <a href="%s" style="display:inline-block;background:#6d4bff;color:#fff;text-decoration:none;padding:10px 18px;border-radius:8px;font-weight:600">%s</a>
  <p style="font-size:12px;color:#888;margin-top:24px">meet.jackdehaan.com — a private availability poll.</p>
</div>`, greeting, bodyHTML, url, ctaLabel)
}
