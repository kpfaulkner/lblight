package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
)

// LBLight is the core of the load balancer.
// Listens to port, parses both headers and request paths and determines (based on configuration) where
// the request should be forwarded on to. All WIP and learning.
type LBLight struct {
	port int

	// match prefix to appropriate router
	pathPrefixToBackendRouter map[string]*BackendRouter

	// match header KEY to a potential router
	headerToBackendRouter map[string]map[string]*BackendRouter

	// all BackendRouters.... just single point of reference for stats gathering.
	allBackendRouters []*BackendRouter

	// listen for TLS traffic (not behind TLS endpoint)
	tlsListener bool

	// just used to lock when we're gathering stats.
	statsMux sync.RWMutex
}

func NewLBLight(port int, tlsListener bool) *LBLight {
	lbl := LBLight{}
	lbl.pathPrefixToBackendRouter = make(map[string]*BackendRouter)
	lbl.headerToBackendRouter = make(map[string]map[string]*BackendRouter)
	lbl.tlsListener = tlsListener
	lbl.port = port
	return &lbl
}

// GetBackendRouterByExactPathPrefix returns the backend router which is registered for the exact
// match of "path". This is more for registration.
func (l *LBLight) GetBackendRouterByExactPathPrefix(path string) (*BackendRouter, error) {

	lowerPath := strings.ToLower(path)
	backend, ok := l.pathPrefixToBackendRouter[lowerPath]
	if ok {
		return backend, nil
	}

	return nil, fmt.Errorf("Unable to find matching backend for path %s", path)
}

// GetBackendRouterByPathPrefix Checks all routers that have been registered for path prefixes and
// searches each registered BackendRouter for a prefix match. This means it's NOT just a map lookup
// but iterating over all of them looking for prefix matches. May need to rethink this a bit.
func (l *LBLight) GetBackendRouterByPathPrefix(path string) (*BackendRouter, error) {
	lowerPath := strings.ToLower(path)
	for prefix, router := range l.pathPrefixToBackendRouter {
		if strings.HasPrefix(lowerPath, prefix) {
			return router, nil
		}
	}

	return nil, fmt.Errorf("Unable to find matching backend for path %s", path)
}

func (l *LBLight) GetBackendRouterByHeader(headerName string, headerValue string) (*BackendRouter, error) {

	headerValues, ok := l.headerToBackendRouter[headerName]
	if ok {
		// have a match for header... now check specific value.
		headerNameAndValueBackend, ok2 := headerValues[headerValue]
		if ok2 {
			return headerNameAndValueBackend, nil
		}
	}

	return nil, fmt.Errorf("Unable to find matching backend for header %s : %s", headerName, headerValue)
}

// checkHealthOfAllBackendRouters loops through all BackendRouters, in turn
// each BackendRouter will check each backend and attempt a connection... to determine
// health.
func (l *LBLight) CheckHealthOfAllBackendRouters() error {

	if l.allBackendRouters != nil {
		for _, ber := range l.allBackendRouters {

			// ignoring error return value.
			// The error will be indicating if the backend is healthy or not, and the Backend itself
			// should be logging if its not healthy. Would just be doubling up on logging here.
			ber.checkHealthOfAllBackends()
		}
	}
	return nil
}

// AddBackendRouter register a BackendRouter to both pathPrefix map and header maps for lookup
// at runtime. If we have multiple, then we'd definitely NOT know who the request
// really should go to. If any of the paths/headers fail for thie BER, then fail them all.
func (l *LBLight) AddBackendRouter(ber *BackendRouter) error {

	// list of all backend routers... just for stats.
	l.allBackendRouters = append(l.allBackendRouters, ber)

	// check if path/header already registered.
	if ber.acceptedPaths != nil {
		for path, _ := range ber.acceptedPaths {
			_, err := l.GetBackendRouterByExactPathPrefix(path)
			if err == nil {
				// no error, we already have something registered!
				return fmt.Errorf("Conflict: Backend path %s already registered", path)
			}
		}
	}

	// check headers.
	if ber.acceptedHeaders != nil {
		for header, val := range ber.acceptedHeaders {
			_, err2 := l.GetBackendRouterByHeader(header, val)
			if err2 == nil {
				// no error, we already have something registered!
				return fmt.Errorf("Conflict: Backend header %s : %s already registered", header, val)
			}
		}
	}

	// register valid paths/headers
	if ber.acceptedPaths != nil {
		for path, _ := range ber.acceptedPaths {
			l.pathPrefixToBackendRouter[path] = ber
		}
	}

	if ber.acceptedHeaders != nil {
		for header, val := range ber.acceptedHeaders {
			//l.headerToBackendRouter[header][val] = ber
			var specificHeaderMap map[string]*BackendRouter
			var ok bool
			specificHeaderMap, ok = l.headerToBackendRouter[header]
			if !ok {
				//headerVal, ok2 := specificHeaderMap[val]
				specificHeaderMap = make(map[string]*BackendRouter)
			}
			specificHeaderMap[val] = ber
		}
	}

	return nil
}

// GetBackendStats just a hacky get stats/connection and logs it.
// will be replaced by prometheus/whatever metrics.
func (l *LBLight) GetBackendStats() error {

	for _, ber := range l.allBackendRouters {
		for _, be := range ber.backends {
			err := be.LogStats()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getBackend.... TODO(kpfaulkner) make real!
// just gets first match for now.
func (l *LBLight) getBackend(req *http.Request) (*Backend, error) {

	// just return first one
	backendRouter, err := l.GetBackendRouterByPathPrefix(req.URL.Path)
	if err != nil {
		return nil, err
	}

	// check if we have a backend for this router... if not, make one.
	backend, err := backendRouter.GetBackend()
	return backend, err

}

// handleRequestsAndRedirect determines which BackendRouter should be used for the incoming request.
func (l *LBLight) handleRequestsAndRedirect(res http.ResponseWriter, req *http.Request) {
	//log.Infof("handleRequestsAndRedirect : %s", req.RequestURI)

	retries := GetRetryFromContext(req)
	if retries > RetryAttempts {
		log.Warningf("Max retries for query, failing: %s %s", req.RemoteAddr, req.URL.Path)
		http.Error(res, "Service not available", http.StatusServiceUnavailable)
		return
	}

	backend, err := l.getBackend(req)
	if err != nil {
		log.Errorf("Unable to find backend for URL %s", req.RequestURI)
		return
	}

	backendConnection, err := backend.GetBackendConnection()
	if err != nil {
		// Assumption (not really valid) that we're under load so we're going to return 429
		log.Errorf("Unable to find backendconnection for URL %s", req.RequestURI)
		res.WriteHeader(http.StatusTooManyRequests)
		return
	}
	defer backendConnection.SetInUse(false) // once finished with connection, then release back to pool.

	backendConnection.ReverseProxy.ServeHTTP(res, req)
	return
}

func (l *LBLight) ListenAndServeTraffic(certCRTPath string, certKeyPath string) error {
	var err error

	// If using behind a TLS termination endpoint (eg Azure LB) then listening for TLS traffic is wrong, since it's already
	// been "stripped" of the TLS encryption at this point.
	if l.tlsListener {
		log.Infof("ListenAndServeTraffic : port %d : crt %s : key %s", l.port, certCRTPath, certKeyPath)
		err = http.ListenAndServeTLS(fmt.Sprintf(":%d", l.port), certCRTPath, certKeyPath, http.HandlerFunc(l.handleRequestsAndRedirect))
	} else {
		err = http.ListenAndServe(fmt.Sprintf(":%d", l.port), http.HandlerFunc(l.handleRequestsAndRedirect))
	}
	if err != nil {
		log.Errorf("SERVER BLEW UP!! %s", err.Error())
	}
	return err
}
