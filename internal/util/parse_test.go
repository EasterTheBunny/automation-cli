package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easterthebunny/automation-cli/internal/util"
)

func TestParseExp(t *testing.T) {
	t.Parallel()

	value := "2e18"

	result, err := util.ParseExp(value)

	require.NoError(t, err)
	assert.Equal(t, uint64(2_000_000_000_000_000_000), result)
}
