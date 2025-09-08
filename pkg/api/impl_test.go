package api

import (
	"bytes"
	"encoding/json"
	"ew/pkg/subscriptions"
	"io"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	id1  = uuid.New()
	id2  = uuid.New()
	id3  = uuid.New()
	repo *subscriptions.SubscriptionRepositoryMock
)

func prepareServ() *fiber.App {
	repo = subscriptions.NewMockRepo([]*subscriptions.Item{
		{ID: id1, ServiceName: "some item", Price: 125, StartDate: time.Now().AddDate(-1, -5, 0), UserId: uuid.New()},
		{ID: id2, ServiceName: "item 2", Price: 250, StartDate: time.Now().AddDate(0, -2, 0), UserId: uuid.New()},
		{ID: id3, ServiceName: "nothing", Price: 500, StartDate: time.Now().AddDate(0, -1, 0), UserId: uuid.New()},
	})
	validate := validator.New()

	server := NewServer(repo, validate)

	webApp := fiber.New()

	RegisterHandlers(webApp, NewStrictHandler(
		server,
		[]StrictMiddlewareFunc{},
	))
	return webApp
}

func TestImplList(t *testing.T) {
	webApp := prepareServ()

	req := httptest.NewRequest("GET", "/subscriptions", nil)

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	title := `some item`
	if !bytes.Contains(body, []byte(title)) {
		t.Errorf("no text found")
	}

	req = httptest.NewRequest("GET", "/subscriptions?limit=1&offset=1", nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	if bytes.Contains(body, []byte(title)) {
		t.Errorf("text found but shouldn`t")
	}

	req = httptest.NewRequest("GET", "/subscriptions?user_id="+repo.Items[0].UserId.String(), nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	if !bytes.Contains(body, []byte(title)) {
		t.Errorf("no text found")
	}
}

func TestImplGet(t *testing.T) {
	webApp := prepareServ()

	req := httptest.NewRequest("GET", "/subscriptions/"+id1.String(), nil)

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	title := `some item`
	if !bytes.Contains(body, []byte(title)) {
		t.Errorf("no text found")
	}
}

func TestImplDelete(t *testing.T) {
	webApp := prepareServ()

	req := httptest.NewRequest("DELETE", "/subscriptions/"+id1.String(), nil)

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 204 {
		t.Errorf("invalid status code: %d, expected %d", resp.StatusCode, 204)
	}

	req = httptest.NewRequest("DELETE", "/subscriptions/"+id1.String(), nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 404 {
		t.Errorf("invalid status code: %d, expected %d", resp.StatusCode, 404)
	}

	i := slices.IndexFunc(repo.Items, func(item *subscriptions.Item) bool {
		return item.ID == id1
	})

	if i != -1 {
		t.Errorf("item wasnt deleted")
	}
}

func TestImplUpdate(t *testing.T) {
	webApp := prepareServ()

	price := 105
	startTime := "11-2000"
	userId := uuid.New()
	item := SubscriptionPatch{
		Price:     &price,
		StartDate: &startTime,
		UserId:    &userId,
	}

	jsonStr, _ := json.Marshal(item)

	req := httptest.NewRequest("PATCH", "/subscriptions/"+id1.String(), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 204 {
		t.Errorf("invalid status code: %d, expected %d", resp.StatusCode, 204)
	}

	i := slices.IndexFunc(repo.Items, func(item *subscriptions.Item) bool {
		return item.ID == id1
	})

	if repo.Items[i].Price != uint(price) {
		t.Errorf("price not changed")
	}
	if repo.Items[i].UserId.String() != userId.String() {
		t.Errorf("user_id not changed")
	}
	start, _ := time.Parse("01-2006", startTime)
	if repo.Items[i].StartDate != start {
		t.Errorf("start_date not changed")
	}
}

func TestImplCreate(t *testing.T) {
	webApp := prepareServ()

	item := Subscription{
		Price:       105,
		StartDate:   "11-2000",
		UserId:      uuid.New(),
		ServiceName: "test",
	}

	jsonStr, _ := json.Marshal(item)

	req := httptest.NewRequest("POST", "/subscriptions/", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d, expected %d", resp.StatusCode, 200)
	}

	body, _ := io.ReadAll(resp.Body)
	respJson := struct {
		SubscriptionId string `json:"subscription_id"`
	}{}
	err := json.Unmarshal(body, &respJson)
	if err != nil {
		t.Errorf("invalid response body: %s", err.Error())
	}

	i := slices.IndexFunc(repo.Items, func(item *subscriptions.Item) bool {
		return item.ID.String() == respJson.SubscriptionId
	})

	if i == -1 {
		t.Errorf("item wasnt created")
	}

	if repo.Items[i].Price != uint(item.Price) {
		t.Errorf("price incorrect")
	}
	if repo.Items[i].UserId.String() != item.UserId.String() {
		t.Errorf("user_id incorrect")
	}
	if repo.Items[i].ServiceName != item.ServiceName {
		t.Errorf("service_name incorrect")
	}
	start, _ := time.Parse("01-2006", item.StartDate)
	if repo.Items[i].StartDate != start {
		t.Errorf("start_date incorrect")
	}
}

func TestImplStats(t *testing.T) {
	webApp := prepareServ()

	start := time.Now().Format("01-2006")
	req := httptest.NewRequest("GET", "/stats?start_date="+start, nil)

	resp, _ := webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	total := `875`
	if !bytes.Contains(body, []byte(total)) {
		t.Errorf("incorrect total, response body: %s", string(body))
	}

	title := `total_price`
	if !bytes.Contains(body, []byte(title)) {
		t.Errorf("no title")
	}

	req = httptest.NewRequest("GET", "/stats?start_date="+start+"&user_id="+repo.Items[0].UserId.String(), nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	total = `125`
	if !bytes.Contains(body, []byte(total)) {
		t.Errorf("incorrect total, response body: %s", string(body))
	}

	req = httptest.NewRequest("GET", "/stats?start_date="+start+"&service_name=nothing", nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	total = `500`
	if !bytes.Contains(body, []byte(total)) {
		t.Errorf("incorrect total, response body: %s", string(body))
	}

	start2 := time.Now().AddDate(0, -1, 0).Format("01-2006")

	req = httptest.NewRequest("GET", "/stats?start_date="+start2+"&user_id="+repo.Items[0].UserId.String(), nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	total = `250`
	if !bytes.Contains(body, []byte(total)) {
		t.Errorf("incorrect total, response body: %s", string(body))
	}

	req = httptest.NewRequest("GET", "/stats?end_date="+start2+"&user_id="+repo.Items[0].UserId.String(), nil)

	resp, _ = webApp.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("invalid status code: %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)

	total = `2125`
	if !bytes.Contains(body, []byte(total)) {
		t.Errorf("incorrect total, response body: %s", string(body))
	}
}
