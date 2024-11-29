package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupRouter(t *testing.T) {
	router := SetupRouter()
	require.NotNil(t, router)
}
