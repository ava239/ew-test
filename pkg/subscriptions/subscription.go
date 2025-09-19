package subscriptions

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID  `db:"id" goqu:"skipinsert"`
	ServiceName string     `db:"service_name"`
	Price       uint       `db:"price"`
	UserId      uuid.UUID  `db:"user_id"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     *time.Time `db:"end_date" goqu:"omitnil"`
}

type PatchItem struct {
	ID          uuid.UUID  `db:"id" goqu:"skipupdate"`
	ServiceName *string    `db:"service_name" goqu:"omitnil"`
	Price       *uint      `db:"price" goqu:"omitnil"`
	UserId      *uuid.UUID `db:"user_id" goqu:"omitnil"`
	StartDate   *time.Time `db:"start_date" goqu:"omitnil"`
	EndDate     *time.Time `db:"end_date" goqu:"omitnil"`
}

type ListParams struct {
	ServiceName *string
	StartDate   *time.Time
	EndDate     *time.Time
	UserId      *uuid.UUID
	Offset      *int
	Limit       *int
}

type ItemRepo interface {
	GetList(context.Context, ListParams) ([]*Item, error)
	GetByID(context.Context, uuid.UUID) (*Item, error)
	Add(context.Context, *Item) (uuid.UUID, error)
	Update(context.Context, *PatchItem) (int64, error)
	Delete(context.Context, uuid.UUID) (int64, error)
}
