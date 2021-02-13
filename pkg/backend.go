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
	url          *url.URL // do we really need this here?
	Alive        bool
	InUse        bool
	mux          sync.RWMutex
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
