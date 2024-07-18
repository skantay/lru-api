package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/skantay/lru-api/internal/cache"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// ILRUCache интерфейс LRU-кэша. Поддерживает только строковые ключи. Поддерживает только простые типы данных в значениях.
type ILRUCache interface {
	// Put запись данных в кэш
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get получение данных из кэша по ключу
	Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error)
	// GetAll получение всего наполнения кэша в виде двух слайсов: слайса ключей и слайса значений. Пары ключ-значения из кэша располагаются на соответствующих позициях в слайсах.
	GetAll(ctx context.Context) (keys []string, values []interface{}, err error)
	// Evict ручное удаление данных по ключу
	Evict(ctx context.Context, key string) (value interface{}, err error)
	// EvictAll ручная инвалидация всего кэша
	EvictAll(ctx context.Context) error
}

type api struct {
	cache ILRUCache
	log   *slog.Logger
}

func New(ILRUCache ILRUCache, log *slog.Logger) http.Handler {
	api := &api{
		cache: ILRUCache,
		log:   log,
	}

	router := chi.NewMux()

	router.Use(middleware.RequestID)
	router.Use(api.logger)
	router.Use(middleware.Recoverer)

	router.Route("/api", func(r chi.Router) {
		r.Get("/lru/{key}", api.get)
		r.Get("/lru", api.getAll)
		r.Post("/lru", api.create)
		r.Delete("/lru/{key}", api.delete)
		r.Delete("/lru", api.flush)
	})

	return router
}

type createRequest struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	TTLSeconds uint        `json:"ttl_seconds"`
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	var request createRequest

	if err := json.Unmarshal(data, &request); err != nil {
		a.log.Debug("bad request", "error", err.Error())

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if request.Key == "" || request.Value == nil {
		a.log.Debug("bad request", "key", request.Key, "value", request.Value)

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if err := a.cache.Put(
		r.Context(),
		request.Key,
		request.Value,
		time.Duration(request.TTLSeconds)*time.Second,
	); err != nil {
		a.log.Error(err.Error())

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}

type getResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt int64       `json:"expires_at"`
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	a.log.Debug(key)

	value, expiresAt, err := a.cache.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, cache.ErrKeyDoesNotExist) {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	response := getResponse{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt.Unix(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		a.log.Error(err.Error())

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type getAllResponse struct {
	Keys   []string      `json:"keys"`
	Values []interface{} `json:"values"`
}

func (a *api) getAll(w http.ResponseWriter, r *http.Request) {
	keys, values, err := a.cache.GetAll(r.Context())
	if err != nil {
		a.log.Error(err.Error())

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(keys) == 0 && len(values) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	response := getAllResponse{
		Keys:   keys,
		Values: values,
	}

	data, err := json.Marshal(response)
	if err != nil {
		a.log.Error(err.Error())

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (a *api) delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	a.log.Debug(key)

	if _, err := a.cache.Evict(r.Context(), key); err != nil {
		if errors.Is(err, cache.ErrKeyDoesNotExist) {
			w.WriteHeader(http.StatusNotFound)

			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) flush(w http.ResponseWriter, r *http.Request) {
	if err := a.cache.EvictAll(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
