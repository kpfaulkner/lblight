package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddBackendSuccess(t *testing.T) {
	ber := NewBackendRouter(nil, make(map[string]bool), BackendRoundRobin)
	be := NewBackend("foo", 1234,1)
	err := ber.AddBackend(be)
	assert.Equal(t, nil, err, "Error not expected")
}

func TestGetBackendFail(t *testing.T) {
	ber := NewBackendRouter(nil, make(map[string]bool), BackendRandom)
	_, err := ber.GetBackend()
	assert.NotNil(t, nil, err, "No backends expected")
}

func TestGetBackendRandomFail(t *testing.T) {
	ber := NewBackendRouter(nil, make(map[string]bool), BackendRoundRobin)
	_, err := ber.GetBackend()
	assert.NotNil(t, nil, err, "No backends expected")
}



