package inmemory

import (
	"context"
	"ew/internal/models/subscriptions"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepository struct {
	Items []*subscriptions.Subscription
}

func (repo *SubscriptionRepository) GetStats(ctx context.Context, params subscriptions.SubscriptionListParams) (int, error) {
	items, err := repo.GetList(ctx, subscriptions.SubscriptionListParams{UserId: params.UserId, ServiceName: params.ServiceName})
	if err != nil {
		return 0, err
	}

	total := 0

	for _, item := range items {
		var (
			startYear, endYear   int
			startMonth, endMonth time.Month
			start, end           time.Time
		)

		if params.StartDate != nil {
			minUnix := math.Max(float64(params.StartDate.Unix()), float64(item.StartDate.Unix()))
			start = time.Unix(int64(minUnix), 0)
		} else {
			start = item.StartDate
		}

		if item.EndDate != nil {
			end = *item.EndDate
		} else {
			end = time.Now()
		}

		if params.EndDate != nil {
			minUnix := math.Min(float64(params.EndDate.Unix()), float64(end.Unix()))
			end = time.Unix(int64(minUnix), 0)
		}

		if start.After(end) {
			start = end.AddDate(0, 1, 0)
		}

		startYear, startMonth, _ = start.Date()
		endYear, endMonth, _ = end.Date()

		diff := int(endMonth-startMonth+1) + 12*(endYear-startYear)

		total += int(item.Price) * diff
	}
	return total, nil
}

func NewRepo(items []*subscriptions.Subscription) *SubscriptionRepository {
	return &SubscriptionRepository{items}
}

func (repo *SubscriptionRepository) GetList(_ context.Context, params subscriptions.SubscriptionListParams) ([]*subscriptions.Subscription, error) {
	items := make([]*subscriptions.Subscription, len(repo.Items))

	if params.StartDate != nil && params.StartDate.After(time.Now()) {
		return []*subscriptions.Subscription{}, nil
	}

	copy(items, repo.Items)

	if params.Offset != nil {
		items = items[*params.Offset:]
	}

	if params.Limit != nil {
		items = items[:*params.Limit]
	}

	if params.UserId != nil {
		items = slices.DeleteFunc(items, func(item *subscriptions.Subscription) bool {
			return item.UserId != *params.UserId
		})
	}

	if params.ServiceName != nil {
		items = slices.DeleteFunc(items, func(item *subscriptions.Subscription) bool {
			return item.ServiceName != *params.ServiceName
		})
	}

	if params.EndDate == nil {
		now := time.Now()
		y, m, _ := now.Date()
		end := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
		params.EndDate = &end
	}

	if params.StartDate != nil {
		items = slices.DeleteFunc(items, func(item *subscriptions.Subscription) bool {
			startBefore := item.StartDate.Before(*params.EndDate)
			endAfter := item.EndDate == nil || item.EndDate.After(*params.StartDate)
			return !startBefore || !endAfter
		})
	} else {
		items = slices.DeleteFunc(items, func(item *subscriptions.Subscription) bool {
			return !item.StartDate.Before(*params.EndDate)
		})
	}

	return items, nil
}

func (repo *SubscriptionRepository) GetByID(_ context.Context, id uuid.UUID) (*subscriptions.Subscription, error) {
	for _, item := range repo.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, subscriptions.NotFound
}

func (repo *SubscriptionRepository) Add(_ context.Context, elem *subscriptions.Subscription) (uuid.UUID, error) {
	elem.ID = uuid.New()
	repo.Items = append(repo.Items, elem)
	return elem.ID, nil
}

func (repo *SubscriptionRepository) Update(_ context.Context, elem *subscriptions.SubscriptionPatch) (int64, error) {
	for _, item := range repo.Items {
		if item.ID == elem.ID {
			if elem.ServiceName != nil {
				item.ServiceName = *elem.ServiceName
			}
			if elem.Price != nil {
				item.Price = *elem.Price
			}
			if elem.UserId != nil {
				item.UserId = *elem.UserId
			}
			if elem.StartDate != nil {
				item.StartDate = *elem.StartDate
			}
			if elem.EndDate != nil {
				item.EndDate = elem.EndDate
			}

			return 1, nil
		}
	}
	return 0, nil
}

func (repo *SubscriptionRepository) Delete(_ context.Context, id uuid.UUID) (int64, error) {
	for i, item := range repo.Items {
		if item.ID == id {
			repo.Items = slices.Delete(repo.Items, i, i+1)
			return 1, nil
		}
	}
	return 0, nil
}
