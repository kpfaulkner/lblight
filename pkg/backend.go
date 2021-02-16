package pkg

import (
	"fmt"
	"sync"
)

// Backend is unique for a given host:port. This might be pointing to a single machine or possibly a LB/cluster.
// The Backend has a collection of BackendConnections. These BackendConnections are the REAL connections to the given
// target machine
type Backend struct {
	Host string
	Port int
	BackendConnections []*BackendConnection
	MaxConnections int
	mux sync.RWMutex

	// Is this backend alive/dead
	Alive    bool
	aliveMux sync.RWMutex

}

func NewBackend(host string, port int, maxConnections int) *Backend {
	be := Backend{}
	be.Host = host
	be.Port = port
	be.Alive = true
	be.MaxConnections = maxConnections
	return &be
}

// GetBackendConnection either retrieves BackendConnection from a pool OR adds new entry to pool (or errors out)
func (ber *Backend) GetBackendConnection() (*BackendConnection, error) {

	// TODO(kpfaulkner) benchmark this!
	ber.mux.Lock()
	defer ber.mux.Unlock()

	// check if we have any backends spare. If so, use it.
	for index, be := range ber.BackendConnections {
		if !be.IsInUse() {
			ber.BackendConnections[index].SetInUse(true)
			return be, nil
		}
	}

	// if none spare but haven't hit maxBackends yet, make one
	if len(ber.BackendConnections) <= ber.MaxConnections {
		bec := NewBackendConnection(fmt.Sprintf("http://%s:%d", ber.Host, ber.Port))
		ber.BackendConnections = append(ber.BackendConnections, bec)
		return bec, nil
	}

	// if cant make any more, return error.
	return nil, fmt.Errorf("unable to provide backendconnection for request")
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
