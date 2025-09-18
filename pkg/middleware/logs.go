package middleware

import (
	"net/http"
	"time"

	"github.com/SenechkaP/subs-tracker/internal/logger"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &WrapperWriter{
			ResponseWriter: w,
			Statuscode:     0,
		}
		next.ServeHTTP(wrapper, r)
		logger.Access.Infof("%d %s %s %v",
			wrapper.Statuscode,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
