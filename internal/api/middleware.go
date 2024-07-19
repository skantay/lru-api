package api

import (
	"net/http"
	"time"
)

func (a *api) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		a.log.Debug("received request", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)

		a.log.Debug("request handled", "path", r.URL.Path, "duration", time.Since(start))
	})

}
