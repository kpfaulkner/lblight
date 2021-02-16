package pkg

import (
	"fmt"
	"math/rand"
	"sync"
)

// BackendRouter is in control of a particular path (eg. /foo). The BackendRouter has a collection of
// Backends. Each Backend is a unique host/ip (could be single machine or another cluster/LB etc).
// The BackendRouter determines which Backend should receive the request, this could be based on
// random/round-robin/load-tracking/wild-guess etc.
type BackendRouter struct {
	// if the beginning of the request is in acceptedPaths, then use this backend.
	acceptedPaths map[string]bool

	// if the header (key) in acceptedHeaders matches the value, then use this backend
	acceptedHeaders map[string]string

	// list of all backends that can be used with the config.
	backends []*Backend

	mux sync.RWMutex
}

func NewBackendRouter(acceptedHeaders map[string]string, acceptedPaths map[string]bool) *BackendRouter {
	ber := BackendRouter{}
	ber.acceptedHeaders = acceptedHeaders
	ber.acceptedPaths = acceptedPaths
	return &ber
}

// AddBackend adds backend to router.
func (ber *BackendRouter) AddBackend(backend *Backend) error {

	ber.mux.Lock()
	defer ber.mux.Unlock()
	ber.backends = append(ber.backends, backend)
	return nil
}

// GetBackend either retrieves backend from a pool OR adds new entry to pool (or errors out)
// This needs to be based on random/load/wild-guess/spirits....
func (ber *BackendRouter) GetBackend() (*Backend, error) {

	// TODO(kpfaulkner) benchmark this!
	ber.mux.Lock()
	defer ber.mux.Unlock()

	// just get random (for now).
	r := rand.Intn(len(ber.backends))
	be := ber.backends[r]
	return be, nil

	/*
		// Just pick the first one for now.
		for _, be := range ber.backends {
			if be.IsAlive() {
				return be, nil
			}
		}  */

	// if cant make any more, return error.
	return nil, fmt.Errorf("unable to provide backend for request")
}
