package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// if you wanna run this test, you need running instance of the server

// func Test_Client(t *testing.T) {
// 	cl, err := New("localhost:6666")
// 	require.Nil(t, err)
// 	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
// 	err = cl.Set(ctx, "foo", "bar")
// 	require.Nil(t, err)
// 	val, err := cl.Get(ctx, "foo")
// 	require.Nil(t, err)
// 	require.Equal(t, val, "bar")
// 	cl.Close()

// }
func Test_WriteMapResp(t *testing.T) {
	m := map[string]string{
		"foo":  "bar",
		"foo2": "bar2",
	}
	cl, err := New("localhost:6666")
	require.Nil(t, err)
	err = cl.Hello(context.Background(), m)
	require.Nil(t, err)
}
