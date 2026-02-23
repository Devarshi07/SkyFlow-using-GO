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
	isUpdate := e.Status == "updated"
	var subject string
	var intro string
	var heading string
	if isUpdate {
		subject = fmt.Sprintf("SkyFlow Booking Updated — %s", e.FlightNumber)
		intro = "Your booking has been updated! Here are your new details:"
		heading = "BOOKING UPDATE"
	} else {
		subject = fmt.Sprintf("SkyFlow Booking Confirmed — %s", e.FlightNumber)
		intro = "Your booking has been confirmed! Here are your details:"
		heading = "BOOKING CONFIRMATION"
	}

	body := fmt.Sprintf(`Dear %s,

%s

══════════════════════════════════════════
  %s
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
`, e.PassengerName, intro, heading, e.BookingID, e.FlightNumber, e.DepartureTime, e.ArrivalTime, e.Seats, e.Amount, e.Status)

	return s.send(e.To, subject, body)
}

func (s *Sender) SendPasswordReset(to, resetLink string) error {
	subject := "SkyFlow — Reset Your Password"

	body := fmt.Sprintf(`Hi there,

We received a request to reset your password for your SkyFlow account.

Click the link below to set a new password:

%s

This link will expire in 30 minutes.

If you didn't request this, you can safely ignore this email.

— The SkyFlow Team
`, resetLink)

	return s.send(to, subject, body)
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
