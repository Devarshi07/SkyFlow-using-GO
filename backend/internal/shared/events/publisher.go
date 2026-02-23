package events

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/logger"
	"github.com/skyflow/skyflow/internal/shared/rabbitmq"
)

// Publisher publishes domain events to RabbitMQ
type Publisher struct {
	mq  *rabbitmq.Client
	log *logger.Logger
}

func NewPublisher(mq *rabbitmq.Client, log *logger.Logger) *Publisher {
	return &Publisher{mq: mq, log: log}
}

func (p *Publisher) PublishPasswordReset(ctx context.Context, email, resetLink string) {
	if p == nil || p.mq == nil {
		return
	}
	evt := PasswordResetEvent{Email: email, ResetLink: resetLink}
	if err := p.mq.Publish(ctx, QueuePasswordReset, evt); err != nil {
		p.log.Error("failed to publish password.reset event", "email", email, "error", err)
	} else {
		p.log.Info("published password.reset event", "email", email)
	}
}

func (p *Publisher) PublishBookingConfirmed(ctx context.Context, evt BookingConfirmedEvent) {
	if p == nil || p.mq == nil {
		return
	}
	if err := p.mq.Publish(ctx, QueueBookingConfirmed, evt); err != nil {
		p.log.Error("failed to publish booking.confirmed event",
			"booking_id", evt.BookingID,
			"error", err,
		)
	} else {
		p.log.Info("published booking.confirmed event",
			"booking_id", evt.BookingID,
			"email", evt.PassengerEmail,
		)
	}
}
