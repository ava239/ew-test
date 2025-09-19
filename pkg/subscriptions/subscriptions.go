package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var NotFound = errors.New("not found")

type SubscriptionRepository struct {
	DB *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{DB: db}
}

func (repo *SubscriptionRepository) GetList(ctx context.Context, params ListParams) ([]*Item, error) {
	items := []*Item{}

	if params.StartDate != nil && params.StartDate.After(time.Now()) {
		return items, nil
	}

	qParts := make([]string, 0, 5)
	args := make([]interface{}, 0, 7)

	qParts = append(qParts, "1=1")

	if params.ServiceName != nil {
		args = append(args, params.ServiceName)
		qParts = append(qParts, fmt.Sprintf("\"service_name\" = $%d", len(args)))
	}

	if params.UserId != nil {
		args = append(args, params.UserId.String())
		qParts = append(qParts, fmt.Sprintf("\"user_id\" = $%d", len(args)))
	}

	if params.EndDate == nil {
		now := time.Now()
		y, m, _ := now.Date()
		end := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
		params.EndDate = &end
	}

	if params.StartDate != nil {
		args = append(args, params.StartDate)
		args = append(args, params.EndDate)
		qParts = append(qParts, fmt.Sprintf("\"start_date\" < $%d and (\"end_date\" is null or \"end_date\" >= $%d)", len(args), len(args)-1))
	} else {
		args = append(args, params.EndDate)
		qParts = append(qParts, fmt.Sprintf("\"start_date\" < $%d", len(args)))
	}

	lo := make([]string, 0, 2)
	if params.Limit != nil {
		args = append(args, params.Limit)
		lo = append(lo, fmt.Sprintf("LIMIT $%d", len(args)))
	}
	if params.Offset != nil {
		args = append(args, params.Offset)
		lo = append(lo, fmt.Sprintf("OFFSET $%d", len(args)))
	}

	q := fmt.Sprintf("SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE %s %s", strings.Join(qParts, " AND "), strings.Join(lo, " "))
	logrus.WithFields(logrus.Fields{"query": q}).Debug("GetList query")

	rows, err := repo.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		subscription := &Item{}
		err = rows.Scan(
			&subscription.ID,
			&subscription.ServiceName,
			&subscription.Price,
			&subscription.UserId,
			&subscription.StartDate,
			&subscription.EndDate,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, subscription)
	}
	return items, nil
}

func (repo *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	subscription := &Item{}
	err := repo.DB.
		QueryRow(ctx, "SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1", id).
		Scan(
			&subscription.ID,
			&subscription.ServiceName,
			&subscription.Price,
			&subscription.UserId,
			&subscription.StartDate,
			&subscription.EndDate,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, NotFound
	}
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (repo *SubscriptionRepository) Add(ctx context.Context, elem *Item) (uuid.UUID, error) {
	var newID uuid.UUID
	err := repo.DB.QueryRow(
		ctx, "INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		elem.ServiceName,
		elem.Price,
		elem.UserId,
		elem.StartDate,
		elem.EndDate,
	).Scan(&newID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return newID, nil
}

func (repo *SubscriptionRepository) Update(ctx context.Context, elem *PatchItem) (int64, error) {
	qParts := make([]string, 0, 5)
	args := make([]interface{}, 0, 6)

	if elem.ServiceName != nil {
		args = append(args, elem.ServiceName)
		qParts = append(qParts, fmt.Sprintf("\"service_name\" = $%d", len(args)))
	}

	if elem.Price != nil {
		args = append(args, elem.Price)
		qParts = append(qParts, fmt.Sprintf("\"price\" = $%d", len(args)))
	}

	if elem.UserId != nil {
		args = append(args, elem.UserId)
		qParts = append(qParts, fmt.Sprintf("\"user_id\" = $%d", len(args)))
	}

	if elem.StartDate != nil {
		args = append(args, elem.StartDate)
		qParts = append(qParts, fmt.Sprintf("\"start_date\" = $%d", len(args)))
	}

	if elem.EndDate != nil {
		args = append(args, elem.EndDate)
		qParts = append(qParts, fmt.Sprintf("\"end_date\" = $%d", len(args)))
	}

	args = append(args, elem.ID)
	q := fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = $%d", strings.Join(qParts, ","), len(args))

	result, err := repo.DB.Exec(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (repo *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := repo.DB.Exec(ctx, "DELETE FROM subscriptions WHERE id = $1", id)

	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
