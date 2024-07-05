package client

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func Test_ADD(t *testing.T) {
	cl, err := New("localhost:6666")
	key := "one"
	require.Nil(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Set(ctx, key, "1")
	require.Nil(t, err)
	val, err := cl.Get(context.Background(), key)
	require.Nil(t, err)
	require.Equal(t, val, "1")
	ctx2, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Add(ctx2, key)
	require.Nil(t, err)
	ctx3, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	val, err = cl.Get(ctx3, key)
	require.Nil(t, err)
	fmt.Println("value")
	fmt.Println(val)
	require.Equal(t, val, "2")
}

func Test_AddN(t *testing.T) {
	cl, err := New("localhost:6666")
	key := "one"
	value := "2"
	require.Nil(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Set(ctx, key, "1")
	require.Nil(t, err)
	val, err := cl.Get(context.Background(), key)
	require.Nil(t, err)
	require.Equal(t, val, "1")
	err = cl.AddN(ctx, key, value)
	require.Nil(t, err)
	val, err = cl.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, val, "3")
}
