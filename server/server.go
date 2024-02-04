package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/mayankshah1607/task-executor/service"
)

type Server struct {
	service         *service.Service
	host            string
	port            string
	gracefulTimeout time.Duration
	log             zerolog.Logger
}

type ServerOpts struct {
	Service               *service.Service
	Host                  string
	Port                  string
	GracfulShutdownPeriod time.Duration
	Logger                zerolog.Logger
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		service: opts.Service,
		host:    opts.Host,
		port:    opts.Port,
		log:     opts.Logger,
	}
}

func (s *Server) Run(ctx context.Context) {
	r := mux.NewRouter()

	r.HandleFunc("/tasks", s.handleTaskSubmission).Methods("POST")
	r.HandleFunc("/tasks", s.handleTaskStatistics).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", s.host, s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	s.log.Info().Str("host", s.host).Str("port", s.port).Msg("Starting HTTP server")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	for {
		// Wait for context to close.
		select {
		case <-ctx.Done():
			// graceful shutdown
			ctx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
			defer cancel()
			srv.Shutdown(ctx)
			s.log.Info().Msg("Shutting down")
			os.Exit(0)
		}
	}
}
