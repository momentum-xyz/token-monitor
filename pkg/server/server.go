package server

import (
	"context"
	"net/http"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ory/x/errorsx"
)

type Server struct {
	tokenNameHandler *TokenNameHandler
}

func NewServer(tokenNameHandler *TokenNameHandler) *Server {
	return &Server{tokenNameHandler: tokenNameHandler}
}

func (s *Server) HTTPHandler() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/token-name", s.tokenNameHandler.NameHandler)
	return r
}

func (s *Server) Start(ctx context.Context) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: s.HTTPHandler(),
	}
	log.Debug("Starting API server on 8080")
	go func() {
		log.Error(errorsx.WithStack(server.ListenAndServe()))
	}()

	go func() {
		<-ctx.Done()
		log.Debug("Closing API server")
		log.Error(server.Shutdown(ctx))
	}()
}
