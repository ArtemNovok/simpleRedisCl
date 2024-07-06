package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseCommmand(t *testing.T) {
	raw := "*4\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$3\r\nbar\r\n$1\r\n0\r\n"
	command, err := ParseCommand(raw)
	require.Nil(t, err)
	fmt.Println(command)
}
