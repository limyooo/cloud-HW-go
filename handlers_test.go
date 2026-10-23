package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---- test helpers ----

func newTestServer(seed ...Item) *Server {
	store := NewStore()
	for _, it := range seed {
		store.Create(it)
	}
	return NewServer(store)
}

func doRequest(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("response is not valid JSON (%q): %v", rec.Body.String(), err)
	}
	return v
}

// ---- GET ----

func TestHealth(t *testing.T) {
	rec := doRequest(t, newTestServer(), http.MethodGet, "/health", "")
	assertStatus(t, rec, http.StatusOK)
}

func TestGetItems(t *testing.T) {
	t.Run("empty store returns 200 and []", func(t *testing.T) {
		rec := doRequest(t, newTestServer(), http.MethodGet, "/items", "")
		assertStatus(t, rec, http.StatusOK)
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", ct)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %s, want []", got)
		}
	})

	t.Run("returns 200 and all items", func(t *testing.T) {
		srv := newTestServer(Item{Name: "a"}, Item{Name: "b"})
		rec := doRequest(t, srv, http.MethodGet, "/items", "")
		assertStatus(t, rec, http.StatusOK)
		if items := decodeBody[[]Item](t, rec); len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}
	})
}

func TestGetItem(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{"existing item returns 200", "/items/1", http.StatusOK},
		{"missing item returns 404", "/items/42", http.StatusNotFound},
		{"non-numeric id returns 400", "/items/abc", http.StatusBadRequest},
		{"zero id returns 400", "/items/0", http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(Item{Name: "pen", Price: 1.5})
			rec := doRequest(t, srv, http.MethodGet, tc.path, "")
			assertStatus(t, rec, tc.want)
			if tc.want == http.StatusOK {
				if it := decodeBody[Item](t, rec); it.ID != 1 || it.Name != "pen" {
					t.Fatalf("got %+v", it)
				}
			}
		})
	}
}

// ---- POST ----

func TestCreateItem(t *testing.T) {
	t.Run("valid body returns 201 with Location header", func(t *testing.T) {
		srv := newTestServer()
		rec := doRequest(t, srv, http.MethodPost, "/items", `{"name":"pen","price":1.5}`)
		assertStatus(t, rec, http.StatusCreated)

		if loc := rec.Header().Get("Location"); loc != "/items/1" {
			t.Fatalf("Location = %q, want /items/1", loc)
		}
		if it := decodeBody[Item](t, rec); it.ID != 1 || it.Name != "pen" || it.Price != 1.5 {
			t.Fatalf("got %+v", it)
		}
		// The new item can be fetched afterwards.
		assertStatus(t, doRequest(t, srv, http.MethodGet, "/items/1", ""), http.StatusOK)
	})

	bad := []struct{ name, body string }{
		{"empty body", ""},
		{"malformed JSON", `{"name":`},
		{"missing name", `{"price":1}`},
		{"negative price", `{"name":"pen","price":-1}`},
		{"unknown field", `{"name":"pen","color":"red"}`},
	}
	for _, tc := range bad {
		t.Run(tc.name+" returns 400", func(t *testing.T) {
			rec := doRequest(t, newTestServer(), http.MethodPost, "/items", tc.body)
			assertStatus(t, rec, http.StatusBadRequest)
		})
	}
}

// ---- PUT ----

func TestUpdateItem(t *testing.T) {
	t.Run("existing item returns 200 and is updated", func(t *testing.T) {
		srv := newTestServer(Item{Name: "old", Price: 1})
		rec := doRequest(t, srv, http.MethodPut, "/items/1", `{"name":"new","price":2}`)
		assertStatus(t, rec, http.StatusOK)
		if it := decodeBody[Item](t, rec); it.ID != 1 || it.Name != "new" || it.Price != 2 {
			t.Fatalf("got %+v", it)
		}

		get := doRequest(t, srv, http.MethodGet, "/items/1", "")
		if it := decodeBody[Item](t, get); it.Name != "new" {
			t.Fatalf("change not persisted, got %+v", it)
		}
	})

	tests := []struct {
		name, path, body string
		want             int
	}{
		{"missing item returns 404", "/items/42", `{"name":"x"}`, http.StatusNotFound},
		{"malformed JSON returns 400", "/items/1", `not json`, http.StatusBadRequest},
		{"missing name returns 400", "/items/1", `{"price":3}`, http.StatusBadRequest},
		{"invalid id returns 400", "/items/abc", `{"name":"x"}`, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(Item{Name: "old"})
			assertStatus(t, doRequest(t, srv, http.MethodPut, tc.path, tc.body), tc.want)
		})
	}
}

// ---- DELETE ----

func TestDeleteItem(t *testing.T) {
	t.Run("existing item returns 204 and is removed", func(t *testing.T) {
		srv := newTestServer(Item{Name: "temp"})
		rec := doRequest(t, srv, http.MethodDelete, "/items/1", "")
		assertStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("204 response should have no body, got %q", rec.Body.String())
		}
		assertStatus(t, doRequest(t, srv, http.MethodGet, "/items/1", ""), http.StatusNotFound)
	})

	t.Run("missing item returns 404", func(t *testing.T) {
		rec := doRequest(t, newTestServer(), http.MethodDelete, "/items/42", "")
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		rec := doRequest(t, newTestServer(), http.MethodDelete, "/items/abc", "")
		assertStatus(t, rec, http.StatusBadRequest)
	})
}

// ---- routing ----

func TestUnsupportedMethodReturns405(t *testing.T) {
	rec := doRequest(t, newTestServer(Item{Name: "x"}), http.MethodPatch, "/items/1", `{"name":"y"}`)
	assertStatus(t, rec, http.StatusMethodNotAllowed)
}

func TestUnknownRouteReturns404(t *testing.T) {
	rec := doRequest(t, newTestServer(), http.MethodGet, "/nope", "")
	assertStatus(t, rec, http.StatusNotFound)
}
