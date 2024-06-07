package httpsrv

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"aerospike.com/rrd/internal/httpsrv/handlers"
)

const defaultTimeout = 15 * time.Second

// Server contains http server with handlers.
type Server struct {
	srv *http.Server
}

// NewServer returns new http server for serving API.
func NewServer(port int, handlers *handlers.RRD) *Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      NewRouter(handlers),
		WriteTimeout: defaultTimeout,
		ReadTimeout:  defaultTimeout,
	}
	return &Server{
		srv: srv,
	}
}

// Start starts http server.
func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// NewRouter registers router paths.
func NewRouter(handlers *handlers.RRD) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/metrics", handlers.Create).Methods("PUT")
	r.HandleFunc("/metrics", handlers.GetByRange).Methods("GET")

	return r
}
