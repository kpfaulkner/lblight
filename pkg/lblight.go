package pkg

import (
	"errors"
	"fmt"
	"net/http/httputil"
	"sync"
)

// Backend has the ReverseProxy to the real backend server.
type Backend struct {
	url          string // do we really need this here?
	Alive        bool
	InUse        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewBackend(url string) *Backend {
	be := Backend{}
	be.url = url
	be.Alive = false
	be.InUse = false
	be.ReverseProxy = httputil.NewSingleHostReverseProxy(url)
	return &be
}

// BackendRouter points to the REAL server doing the work, ie what the LB is connecting to.
// includes list of header values and/or url paths that will be accepted for this backend.
type BackendRouter struct {
	host string
	port int

	// if the beginning of the request is in acceptedPaths, then use this backend.
	acceptedPaths map[string]bool

	// if the header (key) in acceptedHeaders matches the value, then use this backend
	acceptedHeaders map[string]string

	// list of all backends that can be used with the config.
	backends []Backend
}

func NewBackendRouter(host string, port int, acceptedHeaders map[string]string, acceptedPaths map[string]bool) *BackendRouter {
	ber := BackendRouter{}
	ber.host = host
	ber.port = port
	ber.acceptedHeaders = acceptedHeaders
	ber.acceptedPaths = acceptedPaths
	return &ber
}

// LBLight is the core of the load balancer.
// Listens to port, parses both headers and request paths and determines (based on configuration) where
// the request should be forwarded on to. All WIP and learning.
type LBLight struct {

	// match prefix to appropriate router
	pathPrefixToBackendRouter map[string]*BackendRouter

	// match header KEY to a potential router
	headerToBackendRouter map[string]map[string]*BackendRouter
}

func NewLBLight() *LBLight {
	lbl := LBLight{}
	return &lbl
}

// AddBackendRouter register a BackendRouter to both pathPrefix map and header maps for lookup
// at runtime. If we have multiple, then we'd definitely NOT know who the request
// really should go to. If any of the paths/headers fail for thie BER, then fail them all.
func (l *LBLight) AddBackendRouter(ber *BackendRouter) error {

	// check if path/header already registered.
	for path, _ := range ber.acceptedPaths {
		_, ok := l.pathPrefixToBackendRouter[path]
		if ok {
			// conflict. already have something!
			return fmt.Errorf("Conflict: Backend path %s already registered", path)

		}
	}

	// check headers.
	for header, val := range ber.acceptedHeaders {
		existingHeaderMatch, ok := l.headerToBackendRouter[header]
		if ok {
			_, ok2 := existingHeaderMatch[val]
			if ok2 {
				return fmt.Errorf("Conflict: Backend header %s : %s already registered", header, val)
			}
		}
	}

	// register valid paths/headers
	for path, _ := range ber.acceptedPaths {
		l.pathPrefixToBackendRouter[path] = ber
	}

	for header, val := range ber.acceptedHeaders {
		l.headerToBackendRouter[header][val] = ber
	}

	return nil
}
