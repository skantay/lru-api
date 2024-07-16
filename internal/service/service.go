package service

import (
	"context"
	"time"
)

// ILRUCache интерфейс LRU-кэша. Поддерживает только строковые ключи. Поддерживает только простые типы данных в значениях.
type ILRUCache interface {
	// Put запись данных в кэш
	Put(context.Context, string, interface{}, time.Duration) error

	// Get получение данных из кэша по ключу
	Get(context.Context, string) (interface{}, time.Time, error)

	// GetAll получение всего наполнения кэша в виде двух слайсов: слайса ключей и слайса значений. Пары ключ-значения из кэша располагаются на соответствующих позициях в слайсах.
	GetAll(context.Context) ([]string, []interface{}, error)

	// Evict ручное удаление данных по ключу
	Evict(context.Context, string) (interface{}, error)

	// EvictAll ручная инвалидация всего кэша
	EvictAll(context.Context) error
}
