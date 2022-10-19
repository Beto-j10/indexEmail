package app

import "github.com/go-chi/chi/v5"

func (s *Server) routes() *chi.Mux {
	router := s.router
	router.Route("/api/v1", func(router chi.Router) {
		router.Get("/ping", s.ping())
	})
	return router
}
