package workers

import (
	"encoding/json"
	"fmt"

	"github.com/skyflow/skyflow/internal/shared/email"
	"github.com/skyflow/skyflow/internal/shared/events"
	"github.com/skyflow/skyflow/internal/shared/logger"
	"github.com/skyflow/skyflow/internal/shared/rabbitmq"
)

type EmailWorker struct {
	mq     *rabbitmq.Client
	sender *email.Sender
	log    *logger.Logger
}

func NewEmailWorker(mq *rabbitmq.Client, sender *email.Sender, log *logger.Logger) *EmailWorker {
	return &EmailWorker{mq: mq, sender: sender, log: log}
}

// Start declares queues and starts consuming events
func (w *EmailWorker) Start() error {
	// Booking confirmation emails
	if err := w.mq.DeclareQueue(events.QueueBookingConfirmed); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if err := w.mq.Consume(events.QueueBookingConfirmed, "email-worker-booking", func(body []byte) error {
		var evt events.BookingConfirmedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			w.log.Error("failed to unmarshal booking event", "error", err)
			return nil
		}
		w.log.Info("processing booking confirmation email",
			"booking_id", evt.BookingID,
			"email", evt.PassengerEmail,
			"flight", evt.FlightNumber,
		)
		return w.sender.SendBookingConfirmation(email.BookingEmail{
			To:            evt.PassengerEmail,
			PassengerName: evt.PassengerName,
			BookingID:     evt.BookingID,
			FlightNumber:  evt.FlightNumber,
			DepartureTime: evt.DepartureTime,
			ArrivalTime:   evt.ArrivalTime,
			Seats:         evt.Seats,
			Amount:        fmt.Sprintf("$%.2f", float64(evt.AmountCents)/100),
			Status:        evt.Status,
		})
	}); err != nil {
		return fmt.Errorf("consume booking queue: %w", err)
	}

	// Password reset emails
	if err := w.mq.DeclareQueue(events.QueuePasswordReset); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if err := w.mq.Consume(events.QueuePasswordReset, "email-worker-reset", func(body []byte) error {
		var evt events.PasswordResetEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			w.log.Error("failed to unmarshal password reset event", "error", err)
			return nil
		}
		w.log.Info("processing password reset email", "email", evt.Email)
		return w.sender.SendPasswordReset(evt.Email, evt.ResetLink)
	}); err != nil {
		return fmt.Errorf("consume reset queue: %w", err)
	}

	w.log.Info("email worker started",
		"queues", []string{events.QueueBookingConfirmed, events.QueuePasswordReset},
	)
	return nil
}
