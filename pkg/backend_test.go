package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetRetryFromContextSuccess(t *testing.T) {
	req := http.Request{}
	parentCtx := context.TODO()
	ctx := context.WithValue(parentCtx, RetryID,123)
	reqWithContext := req.WithContext(ctx)
	retryVal := GetRetryFromContext(reqWithContext)
	assert.Equal(t, 123, retryVal)
}

func TestGetRetryFromContextWithEmptyContext(t *testing.T) {
	req := http.Request{}
	retryVal := GetRetryFromContext(&req)
	assert.Equal(t, 0, retryVal)
}

func TestGetBackendConnectionNoConnectionsAvailable(t *testing.T) {
	be := NewBackend("myhost", 1234, 0)
	_,err := be.GetBackendConnection()
	assert.NotEqual(t, nil, err,"Expected no connections")
}

func TestGetBackendConnectionConnectionAvailable(t *testing.T) {
	be := NewBackend("myhost", 1234, 1)
	bec,err := be.GetBackendConnection()
	assert.Equal(t, nil, err,"Error!")
	assert.NotEqual(t, nil, bec, "BackendConnection should not be nil")
}

func TestGetBackendConnectionWithExistingConnectionInUse(t *testing.T) {
	be := NewBackend("myhost", 1234, 1)
	bec,err := be.GetBackendConnection()
	assert.Equal(t, nil, err,"Error!")
	assert.NotEqual(t, nil, bec, "BackendConnection should not be nil")

	bec.SetInUse(true)

	// now try and get connection again.
	bec2,err := be.GetBackendConnection()
	assert.NotNil(t, err,"Error!")
	assert.Nil(t, bec2, "Should not get backendconnection")


}
