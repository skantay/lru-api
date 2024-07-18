package cache

import (
	"context"
	"testing"
	"time"

	"github.com/skantay/lru-api/internal/cache/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPut(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value interface{}
		ttl   time.Duration
	}

	tests := []struct {
		name        string
		args        args
		expectedErr error
	}{
		{
			name: "#1: put new string value",
			args: args{
				ctx:   context.Background(),
				key:   "key 1",
				value: "value 1",
				ttl:   0,
			},
		},
		{
			name: "#2: put new int value",
			args: args{
				ctx:   context.Background(),
				key:   "key 2",
				value: 1,
				ttl:   0,
			},
		},
		{
			name: "#3: update key 2",
			args: args{
				ctx:   context.Background(),
				key:   "key 2",
				value: 2,
				ttl:   0,
			},
		},
		{
			name: "#4: put new value and evict key 1(least used)",
			args: args{
				ctx:   context.Background(),
				key:   "key 3",
				value: 2,
				ttl:   0,
			},
		},
		{
			name: "#5: update a node that was not recently used",
			args: args{
				ctx:   context.Background(),
				key:   "key 2",
				value: 3,
				ttl:   0,
			},
		},
	}

	cache, err := New(2, time.Second*60, &mocks.Logger{})
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := cache.Put(
				test.args.ctx,
				test.args.key,
				test.args.value,
				test.args.ttl,
			)

			assert.NoError(t, err)

		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}

	type put struct {
		ctx   context.Context
		key   string
		value interface{}
		ttl   time.Duration
	}

	tests := []struct {
		name           string
		args           args
		put            *put
		expectedValue  interface{}
		expectedExpiry time.Time
		expectedErr    error
	}{
		{
			name: "#1: non existent key",
			args: args{
				ctx: context.Background(),
				key: "key 1",
			},
			expectedValue: nil,
			expectedErr:   ErrKeyDoesNotExist,
		},
		{
			name: "#2: create key 1 with ttl = 1 && get expired key",
			args: args{
				ctx: context.Background(),
				key: "key 1",
			},
			put: &put{
				ctx:   context.Background(),
				ttl:   1,
				key:   "key 1",
				value: 1,
			},
			expectedValue: nil,
			expectedErr:   ErrKeyDoesNotExist,
		},
		{
			name: "#3: create key 1 with ttl = 1000000 && get key 1",
			args: args{
				ctx: context.Background(),
				key: "key 1",
			},
			put: &put{
				ctx:   context.Background(),
				ttl:   1000000,
				key:   "key 1",
				value: 1,
			},
			expectedValue: 1,
			expectedErr:   nil,
		},
		{
			name: "#4: create key 2 && and evict least used",
			args: args{
				ctx: context.Background(),
				key: "key 1",
			},
			put: &put{
				ctx:   context.Background(),
				ttl:   1000000,
				key:   "key 2",
				value: 2,
			},
			expectedValue: nil,
			expectedErr:   ErrKeyDoesNotExist,
		},
		{
			name: "#5: try to get key 2",
			args: args{
				ctx: context.Background(),
				key: "key 2",
			},
			expectedValue: 2,
			expectedErr:   nil,
		},
		{
			name: "#6: again create key 1 && try to get key 2",
			args: args{
				ctx: context.Background(),
				key: "key 2",
			},
			put: &put{
				ctx:   context.Background(),
				ttl:   1000000,
				key:   "key 1",
				value: 1,
			},
			expectedValue: nil,
			expectedErr:   ErrKeyDoesNotExist,
		},
		{
			name: "#7: get key 1",
			args: args{
				ctx: context.Background(),
				key: "key 1",
			},
			expectedValue: 1,
			expectedErr:   nil,
		},
	}

	cache, err := New(1, time.Second*60, &mocks.Logger{})
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.put != nil {
				err := cache.Put(test.put.ctx, test.put.key, test.put.value, test.put.ttl)
				assert.NoError(t, err)
			}
			value, _, err := cache.Get(context.Background(), test.args.key)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedValue, value)
		})
	}
}

func TestGetAll(t *testing.T) {
	cache, err := New(2, time.Second*60, &mocks.Logger{})
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 1", 1, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 2", 2, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 3", 3, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 4", 4, 1000000)
	assert.NoError(t, err)

	keys, values, err := cache.GetAll(context.Background())
	assert.NoError(t, err)

	for i := range keys {
		if keys[i] == "key 1" || values[i] == 1 {
			t.Error("key 1 supposed to be evicted, but found")
		}
	}
}

func TestEvict(t *testing.T) {
	cache, err := New(2, 1000000000000, &mocks.Logger{})
	assert.NoError(t, err)

	value, err := cache.Evict(context.Background(), "key 1")
	assert.Equal(t, ErrKeyDoesNotExist, err)
	assert.Nil(t, value)

	err = cache.Put(context.Background(), "key 1", 1, 1000000)
	assert.NoError(t, err)

	value, err = cache.Evict(context.Background(), "key 1")
	assert.NoError(t, err)
	assert.Equal(t, 1, value)

	keys, values, err := cache.GetAll(context.Background())
	assert.NoError(t, err)

	if len(keys) != 0 && len(values) != 0 {
		t.Error("cache is supposed to be empty")
	}
}

func TestEvictAll(t *testing.T) {
	cache, err := New(2, time.Second*60, &mocks.Logger{})
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 1", 1, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 2", 2, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 3", 3, 1000000)
	assert.NoError(t, err)

	err = cache.Put(context.Background(), "key 4", 4, 1000000)
	assert.NoError(t, err)

	err = cache.EvictAll(context.Background())
	assert.NoError(t, err)

	keys, values, err := cache.GetAll(context.Background())
	assert.NoError(t, err)

	if len(keys) != 0 && len(values) != 0 {
		t.Error("cache is supposed to be empty")
	}
}
