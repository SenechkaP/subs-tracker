package req

import (
	"net/http"
)

func HandleBody[T any](q *http.Request) (*T, error) {
	body, err := Decode[T](q.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}
