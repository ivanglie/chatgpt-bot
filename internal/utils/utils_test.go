package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenHelpMsg(t *testing.T) {
	require.Equal(t, "cmd _- description_\n", GenHelpMsg([]string{"cmd"}, "description"))
}
