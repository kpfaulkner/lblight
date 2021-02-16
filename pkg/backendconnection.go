package pkg

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// BackendConnection has the ReverseProxy to the real backend server.
// There may be many BackendConnections going from the same Backend to the real host/IP. Each BackendConnection
// will be in it's own goroutine (probably).
type BackendConnection struct {
	url      *url.URL // do we really need this here?
	InUse    bool
	inUseMux sync.RWMutex

	ReverseProxy *httputil.ReverseProxy
}

func NewBackendConnection(uri string) *BackendConnection {
	be := BackendConnection{}
	var err error
	be.url, err = url.Parse(uri) // yes, ignoring error for moment... I'm bad. TODO(kpfaulkner)
	if err != nil {
		log.Fatalf("Unable to generate new BackendConnection ....  intentionally dying")
	}

	be.InUse = false
	be.ReverseProxy = httputil.NewSingleHostReverseProxy(be.url)
	be.ReverseProxy.Transport = &http.Transport{DialTLS: dialTLS}
	return &be
}

func (b *BackendConnection) IsInUse() bool {
	var inUse bool
	b.inUseMux.RLock()
	inUse = b.InUse
	b.inUseMux.RUnlock()
	return inUse
}

func (b *BackendConnection) SetInUse(inUse bool) {
	b.inUseMux.Lock()
	b.InUse = inUse
	b.inUseMux.Unlock()
}
