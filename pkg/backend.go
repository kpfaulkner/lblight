package pkg

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	RetryID int = 1

	RetryAttempts  int           = 10
	RetryDelayInMS time.Duration = 50
)

// Backend is unique for a given host:port. This might be pointing to a single machine or possibly a LB/cluster.
// The Backend has a collection of BackendConnections. These BackendConnections are the REAL connections to the given
// target machine
type Backend struct {
	Host               string
	Port               int
	BackendConnections []*BackendConnection
	MaxConnections     int
	mux                sync.RWMutex

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

// LogStats... just a hack to get some data. Log stats (used connections etc).
func (ber *Backend) LogStats() error {

	becInUse := 0
	for _, bec := range ber.BackendConnections {
		if bec.IsInUse() {
			becInUse++
		}
	}
	log.Infof("Backend %s : currently in use %d", ber.Host, becInUse)
	return nil
}

// GetAttemptsFromContext returns the attempts for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(RetryID).(int); ok {
		return retry
	}
	return 0
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
	if len(ber.BackendConnections) < ber.MaxConnections {
		//log.Infof("backend url %s", ber.Host)
		bec := NewBackendConnection(ber.Host)
		bec.SetInUse(true)
		bec.ReverseProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {

			retries := GetRetryFromContext(request)
			if retries < RetryAttempts {
				log.Errorf("Failed query, delaying and retrying: %d : %s", retries, e.Error()) // TODO(kpfaulkner) add retry logic here.
				<-time.After(RetryDelayInMS * time.Millisecond)
				ctx := context.WithValue(request.Context(), RetryID, retries+1)
				bec.ReverseProxy.ServeHTTP(writer, request.WithContext(ctx))
				return
			}

			ber.SetIsAlive(false)
			log.Errorf("Backend <find ID> returned error. Pausing... %s", e.Error()) // TODO(kpfaulkner) add retry logic here.
			writer.WriteHeader(http.StatusTooManyRequests)
		}

		ber.BackendConnections = append(ber.BackendConnections, bec)
		return bec, nil
	}

	// if cant make any more, return error.
	return nil, fmt.Errorf("unable to provide backendconnection for request")
}

// CheckHealth confirms if can talk to host configured for this backend. If cannot, then mark backend as NOT alive.
// Unsure if should do TCP or HTTP. TCP would have less overhead and really just interested if we can connect... surely?
func (b *Backend) checkHealth() error {
	timeout := 3 * time.Second
	u, _ := url.Parse(b.Host)
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	b.SetIsAlive(err == nil)

	if err != nil {
		log.Infof("healthcheck for %s is %v", b.Host, err == nil)
	}
	if conn != nil {
		conn.Close()
	}
	return nil
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
