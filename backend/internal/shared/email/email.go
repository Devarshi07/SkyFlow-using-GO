package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/skyflow/skyflow/internal/shared/logger"
)

type Sender struct {
	host     string
	port     string
	from     string
	password string
	log      *logger.Logger
	enabled  bool
}

func NewSender(log *logger.Logger) *Sender {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	enabled := host != "" && from != "" && password != ""
	if !enabled {
		log.Warn("SMTP not configured, emails will be logged only")
	} else {
		log.Info("SMTP email sender enabled", "host", host, "from", from)
	}
	if port == "" {
		port = "587"
	}

	return &Sender{
		host:     host,
		port:     port,
		from:     from,
		password: password,
		log:      log,
		enabled:  enabled,
	}
}

type BookingEmail struct {
	To             string
	PassengerName  string
	BookingID      string
	FlightNumber   string
	DepartureTime  string
	ArrivalTime    string
	Seats          int
	Amount         string
	Status         string
}

func (s *Sender) SendBookingConfirmation(e BookingEmail) error {
	subject := fmt.Sprintf("SkyFlow Booking Confirmed — %s", e.FlightNumber)

	body := fmt.Sprintf(`Dear %s,

Your booking has been confirmed! Here are your details:

══════════════════════════════════════════
  BOOKING CONFIRMATION
══════════════════════════════════════════

  Booking ID:    %s
  Flight:        %s
  Departure:     %s
  Arrival:       %s
  Seats:         %d
  Total Paid:    %s
  Status:        %s

══════════════════════════════════════════

Thank you for choosing SkyFlow!
We wish you a pleasant journey.

— The SkyFlow Team
`, e.PassengerName, e.BookingID, e.FlightNumber, e.DepartureTime, e.ArrivalTime, e.Seats, e.Amount, e.Status)

	return s.send(e.To, subject, body)
}

func (s *Sender) send(to, subject, body string) error {
	// Always log the email
	s.log.Info("sending email",
		"to", to,
		"subject", subject,
	)

	if !s.enabled {
		s.log.Info("email logged (SMTP not configured)",
			"to", to,
			"subject", subject,
			"body_preview", truncate(body, 200),
		)
		return nil
	}

	msg := strings.Join([]string{
		"From: SkyFlow <" + s.from + ">",
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	addr := s.host + ":" + s.port

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		s.log.Error("failed to send email", "to", to, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	s.log.Info("email sent successfully", "to", to, "subject", subject)
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
