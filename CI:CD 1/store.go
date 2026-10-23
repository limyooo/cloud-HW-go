package main

import (
	"errors"
	"sort"
	"sync"
)

// Item is a simple record kept in the store.
type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ErrNotFound is returned when an item with the given ID does not exist.
var ErrNotFound = errors.New("item not found")

// Store is a thread-safe, in-memory collection of items.
type Store struct {
	mu     sync.RWMutex
	items  map[int]Item
	nextID int
}

// NewStore returns an empty store whose first item will get ID 1.
func NewStore() *Store {
	return &Store{items: make(map[int]Item), nextID: 1}
}

// List returns all items sorted by ID.
func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Get returns the item with the given ID.
func (s *Store) Get(id int) (Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	return it, nil
}

// Create stores a new item, assigns it the next ID, and returns it.
func (s *Store) Create(it Item) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	it.ID = s.nextID
	s.nextID++
	s.items[it.ID] = it
	return it
}

// Update replaces the item with the given ID. The ID itself never changes.
func (s *Store) Update(id int, it Item) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return Item{}, ErrNotFound
	}
	it.ID = id
	s.items[id] = it
	return it, nil
}

// Delete removes the item with the given ID.
func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
