package subscription

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

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
	ErrFetchSubscriptions      = "FAILED TO FETCH SUBSCRIPTION"
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
	router.HandleFunc("GET /users/{user_id}/subscriptions", handler.GetUserSubscriptions())
}

func (handler *SubscriptionHandler) GetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subIDstring := r.PathValue("sub_id")
		subID, err := uuid.Parse(subIDstring)
		if err != nil {
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		sub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
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
				res.JsonDump(w, ErrorResponse{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}
		userID, err := uuid.Parse(body.UserID)
		if err != nil {
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
			return
		}
		startDate, err := parseMonthYear(body.StartDate)
		if err != nil {
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidStartDate}, http.StatusBadRequest)
			return
		}
		var endDate *time.Time
		if body.EndDate != nil && *body.EndDate != "" {
			t, err := parseMonthYear(*body.EndDate)
			if err != nil {
				res.JsonDump(w, ErrorResponse{Error: ErrInvalidEndDate}, http.StatusBadRequest)
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
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		body, err := req.HandleBody[SubscriptionPatchRequest](r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				res.JsonDump(w, ErrorResponse{Error: ErrEmptyBody}, http.StatusBadRequest)
				return
			}
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		existingSub, err := handler.Repository.GetByID(r.Context(), subID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
			res.JsonDump(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
			log.Printf("db error: %v", err)
			return
		}

		if body.PriceRUB != nil {
			existingSub.PriceRUB = *body.PriceRUB
		}

		if body.StartDate != nil {
			startDate, err := parseMonthYear(*body.StartDate)
			if err != nil {
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
					res.JsonDump(w, ErrorResponse{Error: ErrInvalidEndDate}, http.StatusBadRequest)
					return
				}
				existingSub.EndDate = &endDate
			}
		}

		if existingSub.EndDate != nil && existingSub.StartDate.After(*existingSub.EndDate) {
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidDateInterval}, http.StatusBadRequest)
			return
		}

		sub, err := handler.Repository.Update(r.Context(), existingSub)
		if err != nil {
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
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidSubscriptionUUID}, http.StatusBadRequest)
			return
		}
		if err = handler.Repository.Delete(r.Context(), subID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.JsonDump(w, ErrorResponse{Error: ErrSubscriptionNotFound}, http.StatusNotFound)
				return
			}
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
			res.JsonDump(w, ErrorResponse{Error: ErrInvalidUserUUID}, http.StatusBadRequest)
			return
		}
		subList, err := handler.Repository.ListByUser(r.Context(), userID)
		if err != nil {
			res.JsonDump(w, ErrorResponse{Error: ErrFetchSubscriptions}, http.StatusInternalServerError)
			return
		}

		res.JsonDump(w, subList, http.StatusOK)
	}
}
