// Service layer
// So cache level, or controller layer is not responsible for some business logics
// Feels like unnecessary for such a small project, but decided to include service layer
package service

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidKey = errors.New("no key provided")
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

type Service struct {
	cache ILRUCache
}

func (s *Service) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	return s.cache.Put(ctx, key, value, ttl)
}

func (s *Service) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	return s.cache.Get(ctx, key)
}

func (s *Service) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	return s.cache.GetAll(ctx)
}

func (s *Service) Evict(ctx context.Context, key string) (value interface{}, err error) {
	return s.cache.Evict(ctx, key)
}

func (s *Service) EvictAll(ctx context.Context) error {
	return s.EvictAll(ctx)
}
