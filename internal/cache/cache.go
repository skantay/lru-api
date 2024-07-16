package cache

import (
	"context"
	"sync"
	"time"
)

type LRUCache struct {
	m *sync.Mutex
}

func New() *LRUCache {
	return &LRUCache{}
}

func (l *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (l *LRUCache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	return
}

func (l *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	return
}

func (l *LRUCache) Evict(ctx context.Context, key string) (value interface{}, err error) {
	return
}

func (l *LRUCache) EvictAll(ctx context.Context) error {
	return nil
}
