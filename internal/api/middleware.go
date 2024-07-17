package api

import (
	"fmt"
	"net/http"
	"time"
)

func (a *api) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		a.log.Debug(fmt.Sprintf("DEBUG: Received request: %s %s", r.Method, r.URL.Path))

		next.ServeHTTP(w, r)

		a.log.Debug(fmt.Sprintf("DEBUG: Handler for %s took %v", r.URL.Path, time.Since(start)))
	})

}
