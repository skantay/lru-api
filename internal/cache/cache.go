package cache

import (
	"context"
	"errors"
	"log/slog"
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
	defaultTTL  time.Duration
	len, cap    uint
	values      map[string]*node
	m           *sync.Mutex
	most, least *node
	log         *slog.Logger
}

func New(
	cacheSize uint,
	ttl time.Duration,
	log *slog.Logger,
) (*LRUCache, error) {
	if cacheSize == 0 {
		return nil, ErrInvalidCacheSize
	}

	return &LRUCache{
		cap:        cacheSize,
		defaultTTL: ttl,
		m:          &sync.Mutex{},
		values:     make(map[string]*node),
		log:        log,
	}, nil
}

func (l *LRUCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		l.log.Warn(ctx.Err().Error())
		return ctx.Err()
	default:
	}

	l.m.Lock()
	defer l.m.Unlock()

	if ttl == 0 {
		l.log.Debug("default ttl applied", "key", key)
		ttl = l.defaultTTL
	}

	now := time.Now()
	expiration := now.Add(ttl)

	l.log.Debug("node created/updated", "key", key, "created time", now.Format(time.RFC1123), "expiration time", expiration.Format(time.RFC1123))

	if nodeFound, ok := l.values[key]; !ok {
		l.createNode(
			key,
			&node{
				ttl:   expiration,
				value: value,
				next:  l.most,
				key:   key,
			})
		l.log.Debug("creating new node", "key", key)
	} else {
		nodeFound.value = value
		nodeFound.ttl = expiration

		l.updateNode(nodeFound)

		l.log.Debug("node accessed, updated and moved to the front of LRU cache", "key", key)
	}

	return nil
}

func (l *LRUCache) Get(ctx context.Context, key string) (value interface{}, expiresAt time.Time, err error) {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		l.log.Warn(ctx.Err().Error())
		return nil, time.Time{}, ctx.Err()
	default:
	}

	node, ok := l.values[key]
	if !ok {
		l.log.Warn(ErrKeyDoesNotExist.Error(), "key", key)

		return nil, time.Time{}, ErrKeyDoesNotExist
	}

	if time.Now().After(node.ttl) {
		l.evictNode(node)
		l.log.Debug("node expired and has been evicted", "key", node.key)

		return nil, time.Time{}, ErrKeyDoesNotExist
	}

	l.updateNode(node)
	l.log.Debug("node accessed and moved to the front of LRU cache", "key", node.key)

	value = node.value

	expiresAt = node.ttl

	return
}

func (l *LRUCache) GetAll(ctx context.Context) (keys []string, values []interface{}, err error) {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		l.log.Warn(ctx.Err().Error())
		return nil, nil, ctx.Err()
	default:
	}

	now := time.Now()

	for key, node := range l.values {
		if now.After(node.ttl) {
			l.evictNode(node)
			l.log.Debug("node expired and has been evicted", "key", node.key)
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
		l.log.Warn(ctx.Err().Error())
		return nil, ctx.Err()
	default:
	}

	node, ok := l.values[key]
	if !ok {
		l.log.Warn(ErrKeyDoesNotExist.Error(), "key", key)

		return nil, ErrKeyDoesNotExist
	}

	value = node.value

	l.evictNode(node)
	l.log.Debug("node has been evicted", "key", node.key)

	return
}

func (l *LRUCache) EvictAll(ctx context.Context) error {
	l.m.Lock()
	defer l.m.Unlock()

	select {
	case <-ctx.Done():
		l.log.Warn(ctx.Err().Error())
		return ctx.Err()
	default:
	}

	for key := range l.values {
		delete(l.values, key)
	}

	l.len = 0
	l.most = nil
	l.least = nil

	l.log.Debug("cache successfully has been flushed")

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
			l.log.Debug("least used node has been evicted", "key", l.least.key)

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
