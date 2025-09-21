package inmemory

import (
	"context"
	"ew/internal/models/subscriptions"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	id1 = uuid.New()
	id2 = uuid.New()
	id3 = uuid.New()
	id4 = uuid.New()
)

func prepareRepo() *SubscriptionRepository {
	now := time.Now()
	return NewRepo([]*subscriptions.Subscription{
		{ID: id1, ServiceName: "some item", Price: 125, StartDate: time.Now().AddDate(-1, -5, 0), UserId: uuid.New()},
		{ID: id2, ServiceName: "item 2", Price: 250, StartDate: time.Now().AddDate(0, -2, 0), UserId: uuid.New()},
		{ID: id3, ServiceName: "nothing", Price: 500, StartDate: time.Now().AddDate(0, -1, 0), UserId: uuid.New(), EndDate: &now},
		{ID: id4, ServiceName: "variable", Price: 1500, StartDate: time.Now().AddDate(0, -3, 0), UserId: uuid.New()},
	})
}

func TestInMemorySubscriptionRepository_Add(t *testing.T) {
	repo := prepareRepo()

	elem := subscriptions.Subscription{
		Price:       105,
		StartDate:   time.Date(2000, 11, 1, 0, 0, 0, 0, time.UTC),
		UserId:      uuid.New(),
		ServiceName: "test",
	}
	id, err := repo.Add(context.TODO(), &elem)

	if err != nil {
		t.Errorf("Error adding item: %v", err)
	}
	elem.ID = id

	res := slices.ContainsFunc(repo.Items, func(subscription *subscriptions.Subscription) bool {
		return reflect.DeepEqual(subscription, &elem)
	})
	if !res {
		t.Errorf("not added %v", elem)
	}
}

func TestInMemorySubscriptionRepository_GetByID(t *testing.T) {
	repo := prepareRepo()

	elem, err := repo.GetByID(context.TODO(), id1)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(elem, repo.Items[0]) {
		t.Errorf("not found %v", elem)
	}
}

func TestInMemorySubscriptionRepository_GetList(t *testing.T) {
	repo := prepareRepo()

	items, err := repo.GetList(context.TODO(), subscriptions.SubscriptionListParams{UserId: &repo.Items[0].UserId})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(items, []*subscriptions.Subscription{repo.Items[0]}) {
		t.Errorf("not equal %v", items)
	}

	items, err = repo.GetList(context.TODO(), subscriptions.SubscriptionListParams{ServiceName: &repo.Items[1].ServiceName})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(items, []*subscriptions.Subscription{repo.Items[1]}) {
		t.Errorf("not equal %v", items)
	}

	limit, offset := 2, 1
	items, err = repo.GetList(context.TODO(), subscriptions.SubscriptionListParams{Limit: &limit, Offset: &offset})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(items, []*subscriptions.Subscription{repo.Items[1], repo.Items[2]}) {
		t.Errorf("not equal %v", items)
	}

	startDate1, startDate2 := time.Now(), time.Now().AddDate(1, 0, 0)

	items, err = repo.GetList(context.TODO(), subscriptions.SubscriptionListParams{StartDate: &startDate2})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(items, []*subscriptions.Subscription{}) {
		t.Errorf("not equal %v", items)
	}

	items, err = repo.GetList(context.TODO(), subscriptions.SubscriptionListParams{StartDate: &startDate1})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(items, []*subscriptions.Subscription{repo.Items[0], repo.Items[1], repo.Items[3]}) {
		t.Errorf("not equal %v", items)
	}
}

func TestInMemorySubscriptionRepository_Update(t *testing.T) {
	repo := prepareRepo()

	var (
		price     uint = 100
		userId         = uuid.New()
		startDate      = time.Now().AddDate(-2, -5, 0)
		endDate        = time.Now().AddDate(-1, -2, 0)
		name           = "new name"
	)
	affected, err := repo.Update(context.TODO(), &subscriptions.SubscriptionPatch{ID: id1, Price: &price, UserId: &userId, StartDate: &startDate, EndDate: &endDate, ServiceName: &name})

	if err != nil {
		t.Error(err)
	}
	if affected != 1 {
		t.Errorf("affected %v", affected)
	}
	if !reflect.DeepEqual(repo.Items[0], &subscriptions.Subscription{
		ID: id1, ServiceName: name, Price: price, StartDate: startDate, UserId: userId, EndDate: &endDate,
	}) {
		t.Errorf("not updated %v. expected %v", repo.Items[0], &subscriptions.Subscription{
			ID: id1, ServiceName: name, Price: price, StartDate: startDate, UserId: userId, EndDate: &endDate,
		})
	}
}

func TestInMemorySubscriptionRepository_Delete(t *testing.T) {
	repo := prepareRepo()

	affected, err := repo.Delete(context.TODO(), id2)

	if err != nil {
		t.Error(err)
	}

	if affected != 1 {
		t.Errorf("affected %v", affected)
	}

	res := slices.ContainsFunc(repo.Items, func(subscription *subscriptions.Subscription) bool {
		return reflect.DeepEqual(subscription, &repo.Items[1])
	})
	if res {
		t.Errorf("not deleted")
	}

	affected, err = repo.Delete(context.TODO(), id2)

	if err != nil {
		t.Error(err)
	}

	if affected != 0 {
		t.Errorf("affected %v", affected)
	}
}
