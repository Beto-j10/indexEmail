package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	def "server/pkg/definitions"
)

// m type builds a map structure quickly to send to a Responder
type m map[string]interface{}

// wJSON marshals 'response' to JSON and setting the
// Content-Type as application/json.
func wJSON(w http.ResponseWriter, response interface{}, code int) {
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

func (s *Server) indexer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.indexService.Indexer()
		if err != nil {
			log.Printf("Error indexer: %v", err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		m := m{"message": "success"}
		wJSON(w, m, http.StatusOK)
	}
}

func (s *Server) searchMail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if search := r.URL.Query().Get("search"); search == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var err error

		intPage := 1
		if r.URL.Query().Has("page") {
			if intPage, err = strconv.Atoi(r.URL.Query().Get("page")); err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if intPage < 1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}

		intPageSize := 50
		if r.URL.Query().Has("page-size") {
			if intPageSize, err = strconv.Atoi(r.URL.Query().Get("page-size")); err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if intPageSize < 1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}

		if intPageSize*intPage > 30000 {
			m := m{"message": "too many results. max 30000"}
			wJSON(w, m, http.StatusBadRequest)
			return
		}

		query := &def.Query{
			Search:   r.URL.Query().Get("search"),
			Page:     intPage,
			PageSize: intPageSize,
		}

		response, err := s.indexService.SearchMail(query)
		if err != nil {
			log.Printf("Error searching mail: %v", err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		wJSON(w, response, http.StatusOK)
	}
}
