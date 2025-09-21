package subscriptions

import (
	"context"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	GetList(context.Context, SubscriptionListParams) ([]*Subscription, error)
	GetByID(context.Context, uuid.UUID) (*Subscription, error)
	Add(context.Context, *Subscription) (uuid.UUID, error)
	Update(context.Context, *SubscriptionPatch) (int64, error)
	Delete(context.Context, uuid.UUID) (int64, error)
}
