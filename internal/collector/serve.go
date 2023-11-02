package collector

import (
	"fmt"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(apiKey string, db *DB) *chi.Mux {
	svc := NewService(db)

	r := chi.NewRouter()
	r.Use(XAPIKey(apiKey))
	r.Use(DBTxn(db))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/logs/collect", func(w http.ResponseWriter, r *http.Request) {
		ltcs, err := decodeLogsToCollect(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := svc.NewLogsToCollect(r.Context(), ltcs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Post("/logs/stop", func(w http.ResponseWriter, r *http.Request) {
		soe, err := decodeStopOrErr(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := svc.NewStopOrErr(r.Context(), *soe); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Get("/logs/health", func(w http.ResponseWriter, r *http.Request) {
		got, err := svc.Health(r.Context())
		if err != nil || got == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("%d", *got)))
		w.WriteHeader(http.StatusOK)
	})

	return r
}
