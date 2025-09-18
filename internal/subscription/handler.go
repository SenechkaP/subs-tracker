package subscription

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/SenechkaP/subs-tracker/internal/logger"
	"github.com/SenechkaP/subs-tracker/internal/models"
	"github.com/SenechkaP/subs-tracker/pkg/req"
	"github.com/SenechkaP/subs-tracker/pkg/res"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ErrInvalidSubscriptionUUID = "SUBSCRIPTION UUID IS INVALID"
	ErrInvalidUserUUID         = "USER UUID IS INVALID"
	ErrInvalidStartDate        = "START DATE IS INVALID"
	ErrInvalidEndDate          = "END DATE IS INVALID"
	ErrInvalidDateInterval     = "START DATE MUST BE BEFORE OR EQUAL TO END DATE"
	ErrSubscriptionNotFound    = "SUBSCRIPTION WITH PROVIDED UUID DOESN'T EXIST"
	ErrEmptyBody               = "BODY IS EMPTY"
	ErrFetchSubscriptions      = "FAILED TO FETCH SUBSCRIPTIONS"
	ErrMissingParameter        = "MISSING QUERY PARAMETER"
	ErrInvalidParameter        = "QUERY PARAMETER IS INVALID"
)

type SubscriptionHandlerDeps struct {
	Repository *SubscriptionRepository
}

type SubscriptionHandler struct {
	Repository *SubscriptionRepository
}

func NewSubscriptionHandler(router *http.ServeMux, deps *SubscriptionHandlerDeps) {
	handler := SubscriptionHandler{Repository: deps.Repository}
	router.HandleFunc("GET /subscriptions/{sub_id}", handler.GetSubscription())
	router.HandleFunc("POST /subscriptions", handler.CreateSubscription())
	router.HandleFunc("PATCH /subscriptions/{sub_id}", handler.PatchSubscription())
	router.HandleFunc("DELETE /subscriptions/{sub_id}", handler.DeleteSubscription())
	router.HandleFunc("GET /subscriptions/sum", handler.GetSubscriptionsSumByMonth())
	router.HandleFunc("GET /users/{user_id}/subscriptions", handler.GetUserSubscriptions())
}

func (handler *SubscriptionHandler) GetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("sub_id")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			logger.Log.Warnf("GetSubscription invalid uuid sub_id=%s err=%v", subIDstring, err)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		sub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Log.Warnf("GetSubscription not found sub_id=%s", subID.String())
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			logger.Log.Errorf("GetSubscription db error sub_id=%s err=%v", subID.String(), err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, sub, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) CreateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[SubscriptionCreateRequest](r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Log.Warnf("CreateSubscription empty body")
				res.JsonDump(w, ErrorResponse{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			logger.Log.Warnf("CreateSubscription bad request parse body err=%v", err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		userID, err := uuid.Parse(body.UserID)
		if err != nil {
			logger.Log.Warnf("CreateSubscription invalid user uuid user_id=%s", body.UserID)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
			return
		}
		startDate, err := parseMonthYear(body.StartDate)
		if err != nil {
			logger.Log.Warnf("CreateSubscription invalid start date user_id=%s start=%s", userID.String(), body.StartDate)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidStartDate}, http.StatusBadRequest)
			return
		}
		var endDate *time.Time
		if body.EndDate != nil && *body.EndDate != "" {
			t, err := parseMonthYear(*body.EndDate)
			if err != nil {
				logger.Log.Warnf("CreateSubscription invalid start date user_id=%s end=%s", userID.String(), *body.EndDate)
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidEndDate}, http.StatusBadRequest)
				return
			}
			endDate = &t
			if endDate.Before(startDate) {
				logger.Log.Warnf("CreateSubscription invalid interval user_id=%s start=%s end=%s", userID.String(), body.StartDate, *body.EndDate)
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidDateInterval}, http.StatusBadRequest)
				return
			}
		}

		sub := &models.Subscription{
			Service:   body.Service,
			PriceRUB:  body.PriceRUB,
			UserID:    userID,
			StartDate: startDate,
			EndDate:   endDate,
		}
		sub.GenerateNewUUID(handler.Repository.db)

		if err = handler.Repository.Create(r.Context(), sub); err != nil {
			logger.Log.Errorf("CreateSubscription db error user_id=%s service=%s err=%v", userID.String(), body.Service, err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, SubscriptionCreateResponse{SubID: sub.ID.String()}, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) PatchSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("sub_id")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			logger.Log.Warnf("PatchSubscription invalid sub uuid sub_id=%s", subIDstring)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		body, err := req.HandleBody[SubscriptionPatchRequest](r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Log.Warnf("PatchSubscription empty body")
				res.JsonDump(w, ErrorResponse{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			logger.Log.Warnf("PatchSubscription bad request parse body err=%v", err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		existingSub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Log.Warnf("PatchSubscription not found sub_id=%s", subID.String())
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			logger.Log.Errorf("PatchSubscription db error sub_id=%s err=%v", subID.String(), err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		if body.PriceRUB != nil {
			existingSub.PriceRUB = *body.PriceRUB
		}

		if body.StartDate != nil {
			startDate, err := parseMonthYear(*body.StartDate)
			if err != nil {
				logger.Log.Warnf("PatchSubscription invalid start date sub_id=%s start=%s", subID.String(), *body.StartDate)
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidStartDate}, http.StatusBadRequest)
				return
			}
			existingSub.StartDate = startDate
		}

		if body.EndDate != nil {
			if *body.EndDate == "" {
				existingSub.EndDate = nil
			} else {
				endDate, err := parseMonthYear(*body.EndDate)
				if err != nil {
					logger.Log.Warnf("PatchSubscription invalid end date sub_id=%s end=%s", subID.String(), *body.EndDate)
					res.JsonDump(w, ErrorResponse{Error: ErrInvalidEndDate}, http.StatusBadRequest)
					return
				}
				existingSub.EndDate = &endDate
			}
		}

		if existingSub.EndDate != nil && existingSub.EndDate.Before(existingSub.StartDate) {
			logger.Log.Warnf("PatchSubscription invalid interval sub_id=%s start=%s end=%s", subID.String(), *body.StartDate, *body.EndDate)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidDateInterval}, http.StatusBadRequest)
			return
		}

		sub, err := handler.Repository.Update(r.Context(), existingSub)
		if err != nil {
			logger.Log.Errorf("PatchSubscription db error sub_id=%s err=%v", subID.String(), err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, sub, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) DeleteSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("sub_id")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			logger.Log.Warnf("DeleteSubscription invalid sub uuid sub_id=%s", subIDstring)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		if err = handler.Repository.Delete(r.Context(), subID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Log.Warnf("DeleteSubscription not found sub_id=%s", subID.String())
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			logger.Log.Errorf("DeleteSubscription db error sub_id=%s err=%v", subID.String(), err)
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			return
		}
		res.JsonDump(
			w,
			MessageResponse{
				Message: fmt.Sprintf("Subscription with id %v successfully deleted", subID),
			},
			http.StatusOK)
	}
}

func (handler *SubscriptionHandler) GetUserSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDstring := r.PathValue("user_id")
		userID, err := uuid.Parse(userIDstring)
		if err != nil {
			logger.Log.Warnf("GetUserSubscriptions invalid user uuid user_id=%s", userIDstring)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
			return
		}

		q := r.URL.Query()
		offset := 0
		limit := 10

		if offsetStr := q.Get("offset"); offsetStr != "" {
			if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
				offset = v
			} else {
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidParameter}, http.StatusBadRequest)
				return
			}
		}

		if limitStr := q.Get("limit"); limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
				limit = v
			} else {
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidParameter}, http.StatusBadRequest)
				return
			}
		}

		subList, err := handler.Repository.ListByUser(r.Context(), userID, offset, limit)
		if err != nil {
			logger.Log.Errorf("GetUserSubscriptions db error user_id=%s err=%v", userID.String(), err)
			res.JsonDump(w, ErrorResponse{Error: ErrFetchSubscriptions}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, subList, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) GetSubscriptionsSumByMonth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		startParam := q.Get("start")
		endParam := q.Get("end")
		if startParam == "" || endParam == "" {
			logger.Log.Warnf("GetSubscriptionsSumByMonth missing params")
			res.JsonDump(w, ErrorResponse{Error: ErrMissingParameter}, http.StatusBadRequest)
			return
		}

		start, err := parseMonthYear(startParam)
		if err != nil {
			logger.Log.Warnf("GetSubscriptionsSumByMonth invalid start date param=%s err=%v", startParam, err)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidStartDate}, http.StatusBadRequest)
			return
		}
		end, err := parseMonthYear(endParam)
		if err != nil {
			logger.Log.Warnf("GetSubscriptionsSumByMonth invalid end date param=%s err=%v", endParam, err)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidEndDate}, http.StatusBadRequest)
			return
		}

		if end.Before(start) {
			logger.Log.Warnf("GetSubscriptionsSumByMonth invalid interval start=%s end=%s", startParam, endParam)
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidDateInterval}, http.StatusBadRequest)
			return
		}

		var userID *uuid.UUID
		if userParam := q.Get("user_id"); userParam != "" {
			uid, err := uuid.Parse(userParam)
			if err != nil {
				logger.Log.Warnf("GetSubscriptionsSumByMonth invalid user uuid user_id=%s", userParam)
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
				return
			}
			userID = &uid
		}

		var service *string
		if s := q.Get("service"); s != "" {
			service = &s
		}

		sum, err := handler.Repository.SumPriceByMonthRange(r.Context(), start, end, userID, service)
		if err != nil {
			logger.Log.Errorf("GetSubscriptionsSumByMonth db error: %v", err)
			res.JsonDump(w, ErrorResponse{Error: ErrFetchSubscriptions}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, SubscriptionsPriceSumResponse{PriceSum: sum}, http.StatusOK)
	}
}
