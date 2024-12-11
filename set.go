package main

import "sync"

type Set[T comparable] struct {
	m  map[T]struct{}
	mu sync.RWMutex
}

func NewSet[T comparable](cap int) *Set[T] {
	return &Set[T]{
		m:  make(map[T]struct{}, cap),
		mu: sync.RWMutex{},
	}
}

func (s *Set[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, item)
}

func (s *Set[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[item]
	return ok
}

func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}
