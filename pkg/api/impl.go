package api

import (
	"context"
	"errors"
	"ew/pkg/subscriptions"
	"math"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var InternalError = errors.New("internal server error")

type Server struct {
	Repo      subscriptions.ItemRepo
	Validator *validator.Validate
}

func validateDateFormat(fl validator.FieldLevel) bool {
	val := fl.Field().String()

	_, err := time.Parse("01-2006", val)
	if err != nil {
		return false
	}

	return true
}

func NewServer(repo subscriptions.ItemRepo, validate *validator.Validate) Server {
	validate.RegisterValidation("dateFormat", validateDateFormat)

	listRules := map[string]string{
		"Offset": "omitempty,min=0",
		"Limit":  "omitempty,min=1",
	}
	validate.RegisterStructValidationMapRules(listRules, ListSubscriptionsParams{})

	updateRules := map[string]string{
		"Price":     "omitempty,min=1",
		"UserId":    "omitempty",
		"EndDate":   "omitempty,dateFormat",
		"StartDate": "omitempty,dateFormat",
	}
	validate.RegisterStructValidationMapRules(updateRules, UpdateSubscriptionJSONRequestBody{})

	createRules := map[string]string{
		"Price":       "required,min=1",
		"UserId":      "required",
		"EndDate":     "omitempty,dateFormat",
		"StartDate":   "required,dateFormat",
		"ServiceName": "required",
	}
	validate.RegisterStructValidationMapRules(createRules, CreateSubscriptionJSONRequestBody{})

	statsRules := map[string]string{
		"UserId":      "omitempty",
		"EndDate":     "omitempty,dateFormat",
		"StartDate":   "omitempty,dateFormat",
		"ServiceName": "omitempty",
	}
	validate.RegisterStructValidationMapRules(statsRules, StatsSubscriptionsRequestObject{})

	return Server{Repo: repo, Validator: validate}
}

func (s Server) UpdateSubscription(ctx context.Context, request UpdateSubscriptionRequestObject) (UpdateSubscriptionResponseObject, error) {
	err := s.Validator.Struct(request.Body)
	if err != nil {
		logrus.WithError(err).Error("UpdateSubscription validation failed")
		return UpdateSubscription422JSONResponse{Code: 422, Message: err.Error()}, nil
	}

	item := &subscriptions.PatchItem{
		ID:          request.SubscriptionId,
		ServiceName: request.Body.ServiceName,
		UserId:      request.Body.UserId,
	}

	if request.Body.Price != nil {
		price := uint(*request.Body.Price)
		item.Price = &price
	}

	if request.Body.StartDate != nil {
		parse, err := time.Parse("01-2006", *request.Body.StartDate)
		if err != nil {
			logrus.WithError(err).Error("failed to parse start date")
			return UpdateSubscription422JSONResponse{Code: 422, Message: "incorrect start_date format"}, nil
		}
		item.StartDate = &parse
	}

	if request.Body.EndDate != nil {
		parse, err := time.Parse("01-2006", *request.Body.EndDate)
		if err != nil {
			logrus.WithError(err).Error("failed to parse end date")
			return UpdateSubscription422JSONResponse{Code: 422, Message: "incorrect end_date format"}, nil
		}
		item.EndDate = &parse
	}

	updated, err := s.Repo.Update(ctx, item)
	if err != nil {
		logrus.WithError(err).Error("UpdateSubscription failed")
		return nil, InternalError
	}
	if updated == 0 {
		return UpdateSubscription404Response{}, nil
	}
	logrus.Info("updated subscription")

	return UpdateSubscription204Response{}, nil
}

func convertRepoToResponse(item *subscriptions.Item) Subscription {
	var end *string

	if item.EndDate != nil {
		date := item.EndDate.Format("01-2006")
		end = &date
	}

	return Subscription{
		SubscriptionId: &item.ID,
		ServiceName:    item.ServiceName,
		Price:          int(item.Price),
		UserId:         item.UserId,
		StartDate:      item.StartDate.Format("01-2006"),
		EndDate:        end,
	}
}

func (s Server) ListSubscriptions(ctx context.Context, request ListSubscriptionsRequestObject) (ListSubscriptionsResponseObject, error) {
	err := s.Validator.Struct(request.Params)
	if err != nil {
		logrus.WithError(err).Error("ListSubscriptions validation failed")
		return ListSubscriptionsdefaultJSONResponse{Body: Error{Code: 422, Message: err.Error()}, StatusCode: 422}, nil
	}

	params := subscriptions.ListParams{
		Offset:      request.Params.Offset,
		Limit:       request.Params.Limit,
		UserId:      request.Params.UserId,
		ServiceName: request.Params.ServiceName,
	}

	if request.Params.StartDate != nil {
		parse, err := time.Parse("01-2006", *request.Params.StartDate)
		if err != nil {
			logrus.WithError(err).Error("failed to parse start date")
			return ListSubscriptionsdefaultJSONResponse{Body: Error{Code: 422, Message: err.Error()}, StatusCode: 422}, nil
		}
		params.StartDate = &parse
	}

	if request.Params.EndDate != nil {
		parse, err := time.Parse("01-2006", *request.Params.EndDate)
		if err != nil {
			logrus.WithError(err).Error("failed to parse end date")
			return ListSubscriptionsdefaultJSONResponse{Body: Error{Code: 422, Message: err.Error()}, StatusCode: 422}, nil
		}
		params.EndDate = &parse
	}

	items, err := s.Repo.GetList(ctx, params)
	if err != nil {
		logrus.WithError(err).Error("ListSubscriptions failed")
		return nil, InternalError
	}
	logrus.Info("received subscriptions list")

	res := make([]Subscription, 0, len(items))
	for _, item := range items {
		res = append(res, convertRepoToResponse(item))
	}
	logrus.WithFields(logrus.Fields{
		"length": len(res),
	}).Info("prepared subscriptions list")

	return ListSubscriptions200JSONResponse(res), nil
}

func (s Server) ReadSubscription(ctx context.Context, request ReadSubscriptionRequestObject) (ReadSubscriptionResponseObject, error) {
	item, err := s.Repo.GetByID(ctx, request.SubscriptionId)
	if errors.Is(err, subscriptions.NotFound) {
		return ReadSubscription404Response{}, nil
	}
	if err != nil {
		logrus.WithError(err).Error("ReadSubscription failed")
		return nil, InternalError
	}
	logrus.Info("received subscription")

	return ReadSubscription200JSONResponse(convertRepoToResponse(item)), nil
}

func (s Server) CreateSubscription(ctx context.Context, request CreateSubscriptionRequestObject) (CreateSubscriptionResponseObject, error) {
	err := s.Validator.Struct(request.Body)
	if err != nil {
		logrus.WithError(err).Error("CreateSubscription validation failed")
		return CreateSubscription422JSONResponse{Code: 422, Message: err.Error()}, nil
	}

	item := &subscriptions.Item{
		ServiceName: request.Body.ServiceName,
		UserId:      request.Body.UserId,
		Price:       uint(request.Body.Price),
	}

	parse, err := time.Parse("01-2006", request.Body.StartDate)
	if err != nil {
		logrus.WithError(err).Error("failed to parse start date")
		return CreateSubscription422JSONResponse{Code: 422, Message: "incorrect start_date format"}, nil
	}
	item.StartDate = parse

	if request.Body.EndDate != nil {
		parse, err = time.Parse("01-2006", *request.Body.EndDate)
		if err != nil {
			logrus.WithError(err).Error("failed to parse end date")
			return CreateSubscription422JSONResponse{Code: 422, Message: "incorrect end_date format"}, nil
		}
		item.EndDate = &parse
	}

	id, err := s.Repo.Add(ctx, item)
	if err != nil {
		logrus.WithError(err).Error("CreateSubscription failed")
		return nil, InternalError
	}

	logrus.Info("created subscription")

	return CreateSubscription200JSONResponse{&id}, nil
}

func (s Server) DeleteSubscription(ctx context.Context, request DeleteSubscriptionRequestObject) (DeleteSubscriptionResponseObject, error) {
	deleted, err := s.Repo.Delete(ctx, request.SubscriptionId)
	if err != nil {
		logrus.WithError(err).Error("DeleteSubscription failed")
		return nil, InternalError
	}
	if deleted == 0 {
		return DeleteSubscription404Response{}, nil
	}
	logrus.Info("deleted subscription")

	return DeleteSubscription204Response{}, nil
}

func (s Server) StatsSubscriptions(ctx context.Context, request StatsSubscriptionsRequestObject) (StatsSubscriptionsResponseObject, error) {
	err := s.Validator.Struct(request.Params)
	if err != nil {
		logrus.WithError(err).Error("StatsSubscriptions validation failed")
		return StatsSubscriptionsdefaultJSONResponse{Body: Error{Code: 422, Message: err.Error()}, StatusCode: 422}, nil
	}

	items, err := s.Repo.GetList(ctx, subscriptions.ListParams{UserId: request.Params.UserId, ServiceName: request.Params.ServiceName})
	if err != nil {
		logrus.WithError(err).Error("StatsSubscriptions failed")
		return nil, InternalError
	}

	total := 0

	var periodStart, periodEnd *time.Time

	if request.Params.StartDate != nil {
		st, _ := time.Parse("01-2006", *request.Params.StartDate)
		periodStart = &st
	}
	if request.Params.EndDate != nil {
		en, _ := time.Parse("01-2006", *request.Params.EndDate)
		periodEnd = &en
	}
	logrus.WithFields(logrus.Fields{"end": periodEnd, "start": periodStart}).Info("stats subscriptions")

	for _, item := range items {
		var (
			startYear, endYear   int
			startMonth, endMonth time.Month
			start, end           time.Time
		)

		if periodStart != nil {
			minUnix := math.Max(float64(periodStart.Unix()), float64(item.StartDate.Unix()))
			start = time.Unix(int64(minUnix), 0)
		} else {
			start = item.StartDate
		}

		if item.EndDate != nil {
			end = *item.EndDate
		} else {
			end = time.Now()
		}

		if periodEnd != nil {
			minUnix := math.Min(float64(periodEnd.Unix()), float64(end.Unix()))
			end = time.Unix(int64(minUnix), 0)
		}

		if start.After(end) {
			start = end.AddDate(0, 1, 0)
		}

		startYear, startMonth, _ = start.Date()
		endYear, endMonth, _ = end.Date()

		diff := int(endMonth-startMonth+1) + 12*(endYear-startYear)

		logrus.WithFields(logrus.Fields{"diff": diff, "year1": startYear, "year2": endYear, "month1": startMonth, "month2": endMonth}).Info("stats subscriptions")

		total += int(item.Price) * diff
	}

	logrus.Info("received stats")

	return StatsSubscriptions200JSONResponse{TotalPrice: &total}, nil
}
