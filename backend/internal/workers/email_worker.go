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

// Start declares the queue and starts consuming booking confirmation events
func (w *EmailWorker) Start() error {
	if err := w.mq.DeclareQueue(events.QueueBookingConfirmed); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	w.log.Info("email worker started", "queue", events.QueueBookingConfirmed)

	return w.mq.Consume(events.QueueBookingConfirmed, "email-worker", func(body []byte) error {
		var evt events.BookingConfirmedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			w.log.Error("failed to unmarshal booking event", "error", err)
			return nil // don't requeue malformed messages
		}

		w.log.Info("processing booking confirmation email",
			"booking_id", evt.BookingID,
			"passenger", evt.PassengerName,
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
	})
}
