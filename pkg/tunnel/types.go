package tunnel

import (
	"github.com/rancher/remotedialer"
	"net/http"
	"sync"
)

type target struct {
	id     string
	domain string
	target string
}

type transportKey struct {
	id   string
	host string
}

type transportValue struct {
	tranport *http.Transport
	scheme   string
}

type Router struct {
	sync.RWMutex
	transportLock sync.RWMutex

	Server     *remotedialer.Server
	domains    map[string][]target
	clients    map[string][]target
	transports map[transportKey]transportValue
}

type Route struct {
	ID        string
	Scheme    string
	Transport *http.Transport
}

type Client struct {
	Remote string
	// Map of upstream servers /path=http://ip:port
	UpstreamMap map[string]string
	Token       string
}

type RouterStore interface {
	Lookup(req *http.Request) *Route
	Add(req *http.Request) string
	Remove(req *http.Request)
}
