package subscriptions

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var NotFound = errors.New("not found")

type Subscription struct {
	ID          uuid.UUID  `db:"id" goqu:"skipinsert"`
	ServiceName string     `db:"service_name"`
	Price       uint       `db:"price"`
	UserId      uuid.UUID  `db:"user_id"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     *time.Time `db:"end_date" goqu:"omitnil"`
}

type SubscriptionPatch struct {
	ID          uuid.UUID  `db:"id" goqu:"skipupdate"`
	ServiceName *string    `db:"service_name" goqu:"omitnil"`
	Price       *uint      `db:"price" goqu:"omitnil"`
	UserId      *uuid.UUID `db:"user_id" goqu:"omitnil"`
	StartDate   *time.Time `db:"start_date" goqu:"omitnil"`
	EndDate     *time.Time `db:"end_date" goqu:"omitnil"`
}

type SubscriptionListParams struct {
	ServiceName *string
	StartDate   *time.Time
	EndDate     *time.Time
	UserId      *uuid.UUID
	Offset      *int
	Limit       *int
}
