package subscriptions

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID
	ServiceName string
	Price       uint
	UserId      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type PatchItem struct {
	ID          uuid.UUID
	ServiceName *string
	Price       *uint
	UserId      *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
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
