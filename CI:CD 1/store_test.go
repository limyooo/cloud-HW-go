package main

import (
	"errors"
	"sync"
	"testing"
)

func TestNewStoreIsEmpty(t *testing.T) {
	s := NewStore()
	if got := len(s.List()); got != 0 {
		t.Fatalf("new store has %d items, want 0", got)
	}
}

func TestCreateAssignsIncrementingIDs(t *testing.T) {
	s := NewStore()
	a := s.Create(Item{Name: "a"})
	b := s.Create(Item{Name: "b", ID: 99}) // a caller-supplied ID is ignored
	if a.ID != 1 || b.ID != 2 {
		t.Fatalf("got IDs %d and %d, want 1 and 2", a.ID, b.ID)
	}
}

func TestGet(t *testing.T) {
	s := NewStore()
	created := s.Create(Item{Name: "pen", Price: 1.5})

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get(%d) returned error: %v", created.ID, err)
	}
	if got != created {
		t.Fatalf("Get(%d) = %+v, want %+v", created.ID, got, created)
	}

	if _, err := s.Get(999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(999) error = %v, want ErrNotFound", err)
	}
}

func TestListIsSortedByID(t *testing.T) {
	s := NewStore()
	for _, name := range []string{"c", "a", "b"} {
		s.Create(Item{Name: name})
	}
	items := s.List()
	if len(items) != 3 {
		t.Fatalf("List returned %d items, want 3", len(items))
	}
	for i, it := range items {
		if it.ID != i+1 {
			t.Fatalf("items[%d].ID = %d, want %d", i, it.ID, i+1)
		}
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create(Item{Name: "old", Price: 1})

	updated, err := s.Update(created.ID, Item{ID: 50, Name: "new", Price: 2})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.ID != created.ID || updated.Name != "new" || updated.Price != 2 {
		t.Fatalf("Update returned %+v", updated)
	}
	if got, _ := s.Get(created.ID); got != updated {
		t.Fatalf("stored item = %+v, want %+v", got, updated)
	}

	if _, err := s.Update(999, Item{Name: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update(999) error = %v, want ErrNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()
	created := s.Create(Item{Name: "temp"})

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("item still present after Delete")
	}
	if err := s.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second Delete error = %v, want ErrNotFound", err)
	}
}

// Run with `go test -race` to make sure the mutex really protects the map.
func TestConcurrentCreates(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Create(Item{Name: "x"})
		}()
	}
	wg.Wait()
	if got := len(s.List()); got != 100 {
		t.Fatalf("got %d items after 100 concurrent creates, want 100", got)
	}
}
