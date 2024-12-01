package event

import "context"

type Producer interface {
	ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error
}
