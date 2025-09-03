package subscription

import (
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
			res.JsonDump(w, ErrorResponce{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		res.JsonDump(w, sub, http.StatusOK)
	}
}

func (handler *SubscriptionHandler) CreateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[SubscriptionCreateRequest](r)
		if err != nil {
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
