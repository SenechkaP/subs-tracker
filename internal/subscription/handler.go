package subscription

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/SenechkaP/subs-tracker/internal/models"
	"github.com/SenechkaP/subs-tracker/pkg/req"
	"github.com/SenechkaP/subs-tracker/pkg/res"
	"github.com/google/uuid"
)

const (
	ErrInvalidSubscriptionUUID = "SUBSCRIPTION UUID IS INVALID"
	ErrInvalidUserUUID         = "USER UUID IS INVALID"
	ErrInvalidStartDate        = "START DATE IS INVALID"
	ErrInvalidEndDate          = "END DATE IS INVALID"
	ErrInvalidDateInterval     = "START DATE MUST BE BEFORE OR EQUAL TO END DATE"
	ErrSubscriptionNotFound    = "SUBSCRIPTION WITH PROVIDED UUID DOESN'T EXIST"
	ErrEmptyBody               = "BODY IS EMPTY"
)

type SubscriptionHandlerDeps struct {
	Repository *SubscriptionRepository
}

type SubscriptionHandler struct {
	Repository *SubscriptionRepository
}

func NewSubscriptionHandler(router *http.ServeMux, deps *SubscriptionHandlerDeps) {
	handler := SubscriptionHandler{Repository: deps.Repository}
	router.HandleFunc("GET /subscriptions/{uuid}", handler.GetSubscription())
	router.HandleFunc("POST /subscriptions", handler.CreateSubscription())
	router.HandleFunc("PATCH /subscriptions/{uuid}", handler.PatchSubscription())
}

func (handler *SubscriptionHandler) GetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("uuid")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			res.JsonDump(w, ErrorResponce{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		sub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.JsonDump(w, ErrorResponce{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusInternalServerError)
			log.Printf("db error: %v", err)
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
				res.JsonDump(w, ErrorResponce{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		userID, err := uuid.Parse(body.UserID)
		if err != nil {
			res.JsonDump(w, ErrorResponce{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
			return
		}
		startDate, err := parseMonthYear(body.StartDate)
		if err != nil {
			res.JsonDump(w, ErrorResponce{Error: ErrInvalidStartDate}, http.StatusBadRequest)
			return
		}
		var endDate *time.Time
		if body.EndDate != nil && *body.EndDate != "" {
			t, err := parseMonthYear(*body.EndDate)
			if err != nil {
				res.JsonDump(w, ErrorResponce{Error: ErrInvalidEndDate}, http.StatusBadRequest)
				return
			}
			endDate = &t
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
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, SubscriptionCreateResponce{SubID: sub.ID.String()}, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) PatchSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("uuid")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			res.JsonDump(w, ErrorResponce{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		body, err := req.HandleBody[SubscriptionPatchRequest](r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				res.JsonDump(w, ErrorResponce{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		existingSub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.JsonDump(w, ErrorResponce{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusInternalServerError)
			log.Printf("db error: %v", err)
			return
		}

		if body.PriceRUB != nil {
			existingSub.PriceRUB = *body.PriceRUB
		}

		if body.StartDate != nil {
			startDate, err := parseMonthYear(*body.StartDate)
			if err != nil {
				res.JsonDump(w, ErrorResponce{Error: ErrInvalidStartDate}, http.StatusBadRequest)
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
					res.JsonDump(w, ErrorResponce{Error: ErrInvalidEndDate}, http.StatusBadRequest)
					return
				}
				existingSub.EndDate = &endDate
			}
		}

		if existingSub.EndDate != nil && existingSub.StartDate.After(*existingSub.EndDate) {
			res.JsonDump(w, ErrorResponce{Error: ErrInvalidDateInterval}, http.StatusBadRequest)
			return
		}

		sub, err := handler.Repository.Update(r.Context(), existingSub)
		if err != nil {
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, sub, http.StatusOK)
	}
}
