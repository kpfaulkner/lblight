package pkg

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// Backend has the ReverseProxy to the real backend server.
type Backend struct {
	url      *url.URL // do we really need this here?
	Alive    bool
	InUse    bool
	aliveMux sync.RWMutex
	inUseMux sync.RWMutex

	ReverseProxy *httputil.ReverseProxy
}

func NewBackend(uri string) *Backend {
	be := Backend{}
	var err error
	be.url, err = url.Parse(uri) // yes, ignoring error for moment... I'm bad. TODO(kpfaulkner)
	if err != nil {
		log.Fatalf("Unable to generate new backend....  intentionally dying")
	}

	be.Alive = false
	be.InUse = false
	be.ReverseProxy = httputil.NewSingleHostReverseProxy(be.url)
	be.ReverseProxy.Transport = &http.Transport{DialTLS: dialTLS}
	return &be
}

func (b *Backend) IsAlive() bool {
	var alive bool
	b.aliveMux.RLock()
	alive = b.Alive
	b.aliveMux.RUnlock()
	return alive
}

func (b *Backend) SetIsAlive(alive bool) {
	b.aliveMux.Lock()
	b.Alive = alive
	b.aliveMux.Unlock()
}

func (b *Backend) IsInUse() bool {
	var inUse bool
	b.inUseMux.RLock()
	inUse = b.InUse
	b.inUseMux.RUnlock()
	return inUse
}

func (b *Backend) SetInUse(inUse bool) {
	b.inUseMux.Lock()
	b.InUse = inUse
	b.inUseMux.Unlock()
}
