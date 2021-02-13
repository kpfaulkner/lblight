package pkg

import (
	"fmt"
)

// BackendRouter points to the REAL server doing the work, ie what the LB is connecting to.
// includes list of header values and/or url paths that will be accepted for this backend.
type BackendRouter struct {
	host string
	port int

	maxBackends int

	// if the beginning of the request is in acceptedPaths, then use this backend.
	acceptedPaths map[string]bool

	// if the header (key) in acceptedHeaders matches the value, then use this backend
	acceptedHeaders map[string]string

	// list of all backends that can be used with the config.
	backends []*Backend
}

func NewBackendRouter(host string, port int, acceptedHeaders map[string]string, acceptedPaths map[string]bool, maxBackends int) *BackendRouter {
	ber := BackendRouter{}
	ber.host = host
	ber.port = port
	ber.acceptedHeaders = acceptedHeaders
	ber.acceptedPaths = acceptedPaths
	ber.maxBackends = maxBackends
	return &ber
}

// GetBackend either retrieves backend from a pool OR adds new entry to pool (or errors out)
// TODO(kpfaulkner) add locking.
func (ber *BackendRouter) GetBackend() (*Backend, error) {
	// check if we have any backends spare. If so, use it.
	for index, be := range ber.backends {
		if !be.InUse {
			ber.backends[index].InUse = true
			return be, nil
		}
	}

	// if none spare but haven't hit maxBackends yet, make one
	if len(ber.backends) <= ber.maxBackends {
		be := NewBackend(fmt.Sprintf("https://%s:%d", ber.host, ber.port))
		ber.backends = append(ber.backends, be)
		return be, nil
	}

	// if cant make any more, return error.
	return nil, fmt.Errorf("unable to provide backend for request")
}
