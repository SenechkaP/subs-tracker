package middleware

import "net/http"

type WrapperWriter struct {
	http.ResponseWriter
	Statuscode int
}

func (w *WrapperWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.Statuscode = statusCode
}
