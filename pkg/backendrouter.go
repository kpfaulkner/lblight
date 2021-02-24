package pkg

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

type BackendSelectionMethod int

const (
	BackendRoundRobin       BackendSelectionMethod = 1
	BackendInuseConnections BackendSelectionMethod = 2
	BackendRandom           BackendSelectionMethod = 3
)

var BackendSelectionMap = map[string]BackendSelectionMethod{
	"roundrobin":      BackendRoundRobin,
	"inuseconnection": BackendInuseConnections,
	"random":          BackendRandom,
}

func ParseBackendSelectionString(bes string) BackendSelectionMethod {

	var ok bool
	var b BackendSelectionMethod
	b, ok = BackendSelectionMap[strings.ToLower(bes)]
	if !ok {
		// have to have a default of *some* sort.
		return BackendRandom
	}
	return b
}

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

	// roundrobin etc.
	backendSelectionMethod BackendSelectionMethod

	// last backend selected (for round robin)
	lastBackendSelected int

	mux sync.RWMutex
}

func NewBackendRouter(acceptedHeaders map[string]string, acceptedPaths map[string]bool, bes BackendSelectionMethod) *BackendRouter {
	ber := BackendRouter{}
	ber.acceptedHeaders = acceptedHeaders
	ber.acceptedPaths = acceptedPaths
	ber.backendSelectionMethod = bes
	return &ber
}

// AddBackend adds backend to router.
func (ber *BackendRouter) AddBackend(backend *Backend) error {

	ber.mux.Lock()
	defer ber.mux.Unlock()
	ber.backends = append(ber.backends, backend)
	return nil
}

func (ber *BackendRouter) checkHealthOfAllBackends() error {

	for _, be := range ber.backends {

		// ignoring error return value.
		// The error will be indicating if the backend is healthy or not, and the Backend itself
		// should be logging if its not healthy. Would just be doubling up on logging here.
		_ = be.checkHealth()
	}

	return nil
}

// GetBackend either retrieves backend from a pool OR adds new entry to pool (or errors out)
// This needs to be based on random/load/wild-guess/spirits....
func (ber *BackendRouter) GetBackend() (*Backend, error) {

	// TODO(kpfaulkner) benchmark this!
	ber.mux.Lock()
	defer ber.mux.Unlock()

	switch ber.backendSelectionMethod {
	case BackendRandom:

		count := 5

		// get next alive backend.
		for count > 0 {
			r := rand.Intn(len(ber.backends))
			be := ber.backends[r]
			if be.IsAlive() {
				return be, nil
			}
			count--
		}

		// unable to get backend.... throw error.
		return nil, fmt.Errorf("Unable to get backend <TODO figure out identificiation here>")

	case BackendRoundRobin:
		count := 5
		for count > 0 {
			// need to put this in a lock. TODO(kpfaulkner)
			ber.lastBackendSelected++
			if ber.lastBackendSelected >= len(ber.backends) {
				ber.lastBackendSelected = 0
			}

			be := ber.backends[ber.lastBackendSelected]
			if be.IsAlive() {
				return be, nil
			}
			count--
		}

		// unable to get backend.... throw error.
		return nil, fmt.Errorf("Unable to get backend <TODO figure out identificiation here>")

	case BackendInuseConnections:
		// need to calculate based off number of connections etc.....    TODO(kpfaulkner)
		return nil, fmt.Errorf("Not implemented")
	}

	// if cant make any more, return error.
	return nil, fmt.Errorf("unable to provide backend for request")
}
