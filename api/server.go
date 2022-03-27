package api

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"octopus/task-agent/common"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const versionMatcher = "/v{version:[0-9.]+}"

type Server struct {
	cfg           *common.Config
	server        *HTTPServer
	routers       []Router
	routerSwapper *routerSwapper
}

type routerSwapper struct {
	mu     sync.Mutex
	router *mux.Router
}

// ServeHTTP makes the routerSwapper to implement the http.Handler interface.
// entrance of all http requests
func (rs *routerSwapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rs.mu.Lock()
	router := rs.router
	rs.mu.Unlock()
	router.ServeHTTP(w, r)
}

type HTTPServer struct {
	srv *http.Server
	l   net.Listener
}

func (s *HTTPServer) Serve() error {
	return s.srv.Serve(s.l)
}

func NewServer(config *common.Config) *Server {
	logrus.Debugf("server config: %v", config)
	common.ServerConfig = config
	return &Server{cfg: config}
}

func (s *Server) Accept(cfg *common.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Panicf("create Listener failed: %v", err)
	}

	httpServer := &HTTPServer{
		srv: &http.Server{Addr: addr,},
		l:   listener,
	}
	s.server = httpServer
}

func (s *Server) InitRouter(routers ...Router) {
	s.routers = append(s.routers, routers...)

	m := s.createMux()
	s.routerSwapper = &routerSwapper{
		router: m,
	}
}

func (s *Server) createMux() *mux.Router {
	m := mux.NewRouter()

	logrus.Debug("Registering routers")
	for _, apiRouter := range s.routers {
		for _, r := range apiRouter.Routes() {
			f := r.Handler()

			logrus.Debugf("Registering %s, %s", r.Method(), r.Path())
			m.Path(versionMatcher + r.Path()).Methods(r.Method()).Handler(f)
			m.Path(r.Path()).Methods(r.Method()).Handler(f)
		}
	}

	return m
}

func (s *Server) ServeAPI() {
	s.server.srv.Handler = s.routerSwapper
	logrus.Infof("API listen on %s", s.server.l.Addr())
	if err := s.server.Serve(); err != nil {
		logrus.Errorf("start api server failed: %s", err)
	}
}

func (s *Server) Close() {

}
