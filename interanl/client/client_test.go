package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Client(t *testing.T) {
	cl, err := New("localhost:6666")
	require.Nil(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err = cl.Set(ctx, "foo", "bar")
	require.Nil(t, err)
	val, err := cl.Get(ctx, "foo")
	require.Nil(t, err)
	require.Equal(t, val, "bar")

}