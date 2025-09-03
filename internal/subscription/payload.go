package subscription

type SubscriptionCreateRequest struct {
	Service   string  `json:"service_name"`
	PriceRUB  int64   `json:"price"`
	UserID    string  `json:"user_id"`
	StartDate string  `json:"start_date"`
	EndDate   *string `json:"end_date,omitempty"`
}

type SubscriptionCreateResponce struct {
	SubID string `json:"subscription_id"`
}

type ErrorResponce struct {
	Error string `json:"error"`
}
