package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Server exposes the Store over HTTP.
type Server struct {
	store *Store
	mux   *http.ServeMux
}

func NewServer(store *Store) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	// Go 1.22+ ServeMux matches "METHOD /path/{param}" patterns and
	// automatically answers 405 Method Not Allowed for other verbs.
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /items", s.listItems)
	s.mux.HandleFunc("GET /items/{id}", s.getItem)
	s.mux.HandleFunc("POST /items", s.createItem)
	s.mux.HandleFunc("PUT /items/{id}", s.updateItem)
	s.mux.HandleFunc("DELETE /items/{id}", s.deleteItem)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// GET /health -> 200
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /items -> 200
func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

// GET /items/{id} -> 200 | 400 | 404
func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	it, err := s.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, it)
}

// POST /items -> 201 | 400
func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeItem(w, r)
	if !ok {
		return
	}
	created := s.store.Create(in)
	w.Header().Set("Location", fmt.Sprintf("/items/%d", created.ID))
	writeJSON(w, http.StatusCreated, created)
}

// PUT /items/{id} -> 200 | 400 | 404
func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decodeItem(w, r)
	if !ok {
		return
	}
	updated, err := s.store.Update(id, in)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DELETE /items/{id} -> 204 | 400 | 404
func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := s.store.Delete(id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func decodeItem(w http.ResponseWriter, r *http.Request) (Item, bool) {
	var in Item
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return Item{}, false
	}
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return Item{}, false
	}
	if in.Price < 0 {
		writeError(w, http.StatusBadRequest, "price must be >= 0")
		return Item{}, false
	}
	return in, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
