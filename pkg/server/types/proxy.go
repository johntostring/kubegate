package types

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ConnectionType int

const (
	DirectProxy ConnectionType = 0
	TunnelProxy ConnectionType = 1
)

type Proxy struct {
	Name      string
	ProxyType ConnectionType
	Host      string
	Schema    string
	Transport *http.Transport
}

type Store interface {
	Lookup(name string) (*Proxy, bool)
	AddDirectProxy(name string, url string) error
	AddClientTlsCertDirectProxy(name string, url string, clientTlsKeyPair *tls.Certificate) error
	AddTunnelProxy(name string, url string) error
	AddClientTlsCertProxy(proxyType ConnectionType, name string, url string, clientTlsKeyPair *tls.Certificate) error
	Remove(name string)
}

type InMemoryStore struct {
	lock          sync.RWMutex
	inMemoryStore map[string]*Proxy
}

func (r *InMemoryStore) Lookup(name string) (*Proxy, bool) {
	proxy, ok := r.inMemoryStore[name]
	return proxy, ok
}

func (r *InMemoryStore) AddDirectProxy(name string, url string) error {
	return r.AddClientTlsCertProxy(DirectProxy, name, url, nil)
}

func (r *InMemoryStore) AddTunnelProxy(name string, url string) error {
	return r.AddClientTlsCertProxy(TunnelProxy, name, url, nil)
}

func (r *InMemoryStore) AddClientTlsCertDirectProxy(name string, url string, clientTlsKeyPair *tls.Certificate) error {
	return r.AddClientTlsCertProxy(DirectProxy, name, url, clientTlsKeyPair)
}

func (r *InMemoryStore) AddClientTlsCertProxy(proxyType ConnectionType, name string, url string, clientTlsKeyPair *tls.Certificate) error {
	if r.inMemoryStore == nil {
		r.inMemoryStore = make(map[string]*Proxy)
	}

	schemaHost := strings.Split(url, "://")
	if len(schemaHost) != 2 {
		return fmt.Errorf("invalid url")
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	var transport *http.Transport
	if clientTlsKeyPair != nil {
		transport = DefaultTlsClientTransport(*clientTlsKeyPair)
	} else {
		transport = DefaultTransport()
	}
	ap := &Proxy{
		Name:      name,
		ProxyType: proxyType,
		Host:      schemaHost[1],
		Schema:    schemaHost[0],
		Transport: transport,
	}
	r.inMemoryStore[name] = ap
	return nil
}

func (r *InMemoryStore) Remove(name string) {
	delete(r.inMemoryStore, name)
}

func DefaultTlsClientTransport(certificate tls.Certificate) *http.Transport {
	transport := DefaultTransport()
	transport.TLSClientConfig.Certificates = []tls.Certificate{certificate}
	return transport
}

func DefaultTransport() *http.Transport {
	transport := &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return transport
}

func (p Proxy) GetTransport() *http.Transport {
	return p.Transport
}
func (p Proxy) Error(res http.ResponseWriter, req *http.Request, err error) {
	res.Write([]byte(err.Error()))
	res.WriteHeader(http.StatusInternalServerError)
}
