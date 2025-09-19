package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var NotFound = errors.New("not found")

type SubscriptionRepository struct {
	DB *pgxpool.Pool
	QB goqu.DialectWrapper
}

func NewRepo(db *pgxpool.Pool, qb goqu.DialectWrapper) *SubscriptionRepository {
	return &SubscriptionRepository{DB: db, QB: qb}
}

func (repo *SubscriptionRepository) GetList(ctx context.Context, params ListParams) ([]*Item, error) {
	items := []*Item{}

	if params.StartDate != nil && params.StartDate.After(time.Now()) {
		return items, nil
	}

	query := repo.QB.From("subscriptions").
		Select("id", "service_name", "price", "user_id", "start_date", "end_date")

	if params.ServiceName != nil {
		query = query.Where(goqu.Ex{"service_name": params.ServiceName})
	}

	if params.UserId != nil {
		query = query.Where(goqu.Ex{"user_id": params.UserId})
	}

	if params.EndDate == nil {
		now := time.Now()
		y, m, _ := now.Date()
		end := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
		params.EndDate = &end
	}

	if params.StartDate != nil {
		query = query.Where(goqu.And(
			goqu.C("start_date").Lt(params.EndDate),
			goqu.Or(
				goqu.C("end_date").IsNull(),
				goqu.C("end_date").Gte(params.StartDate),
			),
		))
	} else {
		query = query.Where(goqu.C("start_date").Lt(params.EndDate))
	}

	if params.Limit != nil {
		query = query.Limit(uint(*params.Limit))
	}
	if params.Offset != nil {
		query = query.Offset(uint(*params.Offset))
	}

	q, args, _ := query.Prepared(true).ToSQL()
	logrus.WithFields(logrus.Fields{"query": q, "args": args}).Debug("GetList query")

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

	query := repo.QB.From("subscriptions").
		Select("id", "service_name", "price", "user_id", "start_date", "end_date").
		Where(goqu.Ex{"id": id})

	q, args, _ := query.Prepared(true).ToSQL()
	logrus.WithFields(logrus.Fields{"query": q, "args": args}).Debug("GetByID query")

	err := repo.DB.
		QueryRow(ctx, q, args...).
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

	query := repo.QB.Insert("subscriptions").
		Rows(elem).
		Returning("id")

	q, args, _ := query.Prepared(true).ToSQL()
	logrus.WithFields(logrus.Fields{"query": q, "args": args}).Debug("Add query")

	err := repo.DB.QueryRow(ctx, q, args...).Scan(&newID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return newID, nil
}

func (repo *SubscriptionRepository) Update(ctx context.Context, elem *PatchItem) (int64, error) {
	query := repo.QB.Update("subscriptions").
		Where(goqu.Ex{"id": elem.ID}).
		Set(elem)

	q, args, _ := query.Prepared(true).ToSQL()
	logrus.WithFields(logrus.Fields{"query": q, "args": args}).Debug("Update query")

	result, err := repo.DB.Exec(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (repo *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	query := repo.QB.Delete("subscriptions").
		Where(goqu.Ex{"id": id})

	q, args, _ := query.Prepared(true).ToSQL()
	logrus.WithFields(logrus.Fields{"query": q, "args": args}).Debug("Delete query")

	result, err := repo.DB.Exec(ctx, q, args...)

	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
