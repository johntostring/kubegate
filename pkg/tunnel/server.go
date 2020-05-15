package tunnel

import (
	"crypto/subtle"
	"github.com/rancher/remotedialer"
	"net/http"
)

const (
	IdHeader       = "x-kubegate-tunnel-id"
	UpstreamHeader = "x-kubegate-tunnel-upstream"
)

type Server struct {
	Token  string
	router Router
	server *remotedialer.Server

	DisableWrapTransport bool
}

func NewTunnelServer(token string) *Server {
	s := &Server{
		Token:                token,
		DisableWrapTransport: false,
	}
	s.server = remotedialer.New(s.authorized, remotedialer.DefaultErrorWriter)
	s.router.Server = s.server
	return s
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.server.ServeHTTP(res, req)
}

func (s *Server) Lookup(name string) *Route {
	return s.router.Lookup(name)
}
func (s *Server) AddRouter(req *http.Request) string {
	return s.router.Add(req)
}

func (s *Server) RemoveRouter(req *http.Request) {
	s.router.Remove(req)
}

func (s *Server) tokenValid(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	return subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+s.Token)) == 1
}

func (s *Server) authorized(req *http.Request) (id string, ok bool, err error) {
	defer func() {
		if id == "" {
			ok = false
		}
		if !ok || err != nil {
			req.Header.Del(IdHeader)
		}
	}()

	if !s.tokenValid(req) {
		return "", false, nil
	}

	return s.AddRouter(req), true, nil
}

func (s Server) Error(w http.ResponseWriter, req *http.Request, err error) {
	remotedialer.DefaultErrorWriter(w, req, http.StatusInternalServerError, err)
}
