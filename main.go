package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type noteRequest struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func notesHandler(c *InfraClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in noteRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.ID == "" || strings.TrimSpace(in.Text) == "" {
			http.Error(w, "id and text are required", http.StatusBadRequest)
			return
		}
		emb, err := c.embed(in.Text)
		if err == nil {
			err = c.store(in.ID, in.Text, emb)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": in.ID, "status": "indexed"})
	}
}

func searchHandler(c *InfraClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			http.Error(w, "q is required", http.StatusBadRequest)
			return
		}
		k := 3
		if raw := r.URL.Query().Get("top_k"); raw != "" {
			if n, e := strconv.Atoi(raw); e == nil && n > 0 {
				k = n
			}
		}
		emb, err := c.embed(q)
		if err == nil {
			var hits []string
			hits, err = c.search(emb, k)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"query": q, "results": hits})
				return
			}
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func main() {
	c, err := NewInfraClient()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/notes", notesHandler(c))
	mux.Handle("/search", searchHandler(c))
	http.ListenAndServe(":8080", mux)
}
