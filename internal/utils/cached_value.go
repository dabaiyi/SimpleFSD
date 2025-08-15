// Package service
package utils

import (
	"sync"
	"time"
)

type CachedValue[T any] struct {
	generateTime time.Time
	cachedData   *T
	mu           sync.RWMutex
	cachedTime   time.Duration
	getter       func() *T
}

func NewCachedValue[T any](cachedTime time.Duration, getter func() *T) *CachedValue[T] {
	return &CachedValue[T]{time.Now(), nil, sync.RWMutex{}, cachedTime, getter}
}

func (cachedValue *CachedValue[T]) GetValue() *T {
	cachedValue.mu.RLock()
	if cachedValue.cachedData != nil && time.Since(cachedValue.generateTime) <= cachedValue.cachedTime {
		defer cachedValue.mu.RUnlock()
		return cachedValue.cachedData
	}
	cachedValue.mu.RUnlock()

	cachedValue.mu.Lock()
	defer cachedValue.mu.Unlock()

	if cachedValue.cachedData != nil && time.Since(cachedValue.generateTime) <= cachedValue.cachedTime {
		return cachedValue.cachedData
	}

	cachedValue.cachedData = cachedValue.getter()
	cachedValue.generateTime = time.Now()

	return cachedValue.cachedData
}
