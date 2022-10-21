package app

import (
	"encoding/json"
	"log"
	"net/http"
)

// m type builds a map structure quickly to send to a Responder
type m map[string]interface{}

// wJSON marshals 'm' to JSON and setting the
// Content-Type as application/json.
func wJSON(w http.ResponseWriter, response m, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	r, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error Marshal: %v", err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(r)
}

// pingHandler is a simple health check
func (s *Server) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := m{"message": "success"}
		wJSON(w, m, http.StatusOK)
	}
}

func (s *Server) indexMail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.indexService.IndexMail()
		if err != nil {
			log.Printf("Error indexing mail: %v", err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		m := m{"message": "success"}
		wJSON(w, m, http.StatusOK)
	}
}
