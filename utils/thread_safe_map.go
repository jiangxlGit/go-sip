package utils

import (
	"sync"
)

// Go 的泛型线程安全 map（1.18+）
type ThreadSafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// 初始化
func NewThreadSafeMap[K comparable, V any]() *ThreadSafeMap[K, V] {
	return &ThreadSafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Get（加读锁）
func (m *ThreadSafeMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Set（加写锁）
func (m *ThreadSafeMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}