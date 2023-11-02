package collector

import (
	"context"
	"net/http"
)

const (
	XAPIHeader = "X-API-KEY"
)

func XAPIKey(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      gotKey := r.Header.Get(XAPIHeader)

      if gotKey == "" {
				http.Error(w, "api key not present", http.StatusUnauthorized)
				return
			}
  

			if gotKey != key {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func DBTxn(db *DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := db.db.Begin()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer func() {
				if err != nil {
					tx.Rollback()
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				} else {
					if err = tx.Commit(); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}()

			context := context.WithValue(r.Context(), Txn, tx)
			r = r.WithContext(context)

			next.ServeHTTP(w, r)
		})
	}
}
