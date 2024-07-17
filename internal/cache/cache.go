package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyDoesNotExist  = errors.New("key does not exist")
	ErrInvalidCacheSize = errors.New("invalid cache size")
)

type node struct {
	prev, next *node
	value      interface{}
	key        string
	ttl        time.Time
}

type LRUCache struct {
	defaultTTL  int64
	len, cap    uint
	values      map[string]*node
	m           *sync.Mutex
	most, least *node
}

func New(
	cacheSize uint,
	ttl int64,
) (*LRUCache, error) {
	if cacheSize == 0 {
		return nil, ErrInvalidCacheSize
	}

	return &LRUCache{
		cap:        cacheSize,
		defaultTTL: ttl,
		m:          &sync.Mutex{},
		values:     make(map[string]*node),
	}, nil
}

func (l *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	l.m.Lock()
	defer l.m.Unlock()

	if ttl == 0 {
		ttl = time.Duration(l.defaultTTL)
	}

	expiration := time.Now().Add(ttl)

	if nodeFound, ok := l.values[key]; !ok {
		l.createNode(
			key,
			&node{
				ttl:   expiration,
				value: value,
				next:  l.most,
				key:   key,
			})
	} else {
		nodeFound.value = value
		nodeFound.ttl = expiration

		l.updateNode(nodeFound)
	}

	return nil
}

func (l *LRUCache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		return nil, time.Time{}, ctx.Err()
	default:
	}

	node, ok := l.values[key]
	if !ok {
		return nil, time.Time{}, ErrKeyDoesNotExist
	}

	if time.Now().After(node.ttl) {
		l.evictNode(node)
		return nil, time.Time{}, ErrKeyDoesNotExist
	}

	l.updateNode(node)

	value = node.value

	return
}

func (l *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	now := time.Now()

	for key, node := range l.values {
		if now.After(node.ttl) {
			l.evictNode(node)
		} else {
			keys = append(keys, key)
			values = append(values, node.value)
		}
	}

	return
}
func (l *LRUCache) Evict(ctx context.Context, key string) (value interface{}, err error) {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	node, ok := l.values[key]
	if !ok {
		return nil, ErrKeyDoesNotExist
	}

	value = node.value

	l.evictNode(node)

	return
}

func (l *LRUCache) EvictAll(ctx context.Context) error {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for key := range l.values {
		delete(l.values, key)
	}

	l.len = 0
	l.most = nil
	l.least = nil

	return nil
}

func (l *LRUCache) updateNode(node *node) {
	if node != l.most {
		if node.prev != nil {
			node.prev.next = node.next
		}
		if node.next != nil {
			node.next.prev = node.prev
		}
		if node == l.least {
			l.least = node.prev
		}

		node.prev = nil
		node.next = l.most
		if l.most != nil {
			l.most.prev = node
		}
		l.most = node
	}
}

func (l *LRUCache) evictNode(node *node) {
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}

	if node == l.most {
		l.most = node.next
	}
	if node == l.least {
		l.least = node.prev
	}

	delete(l.values, node.key)
	l.len--
}

func (l *LRUCache) createNode(key string, node *node) {
	if l.most != nil {
		l.most.prev = node
	}

	l.most = node

	if l.least == nil {
		l.least = node
	}

	if l.len == l.cap {
		if l.least != nil {
			leastPrev := l.least.prev

			delete(l.values, l.least.key)

			l.least = leastPrev

			if l.least != nil {
				l.least.next = nil
			}
		}
	} else {
		l.len++
	}

	l.values[key] = node
}
