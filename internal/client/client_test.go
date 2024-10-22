package client

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// if you wanna run this tests, you need running instance of the server
// on right address (localhost:6666), or change address manually in every test
var ctx context.Context = context.Background()

func Test_CLient3(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key := "one"
	value := "value_one"
	ind := 0
	want := []string{}
	for i := 0; i < 6; i++ {
		err = cl.LPush(ctx, key, value, ind)
		require.Nil(t, err)
		want = append(want, value)
		isExist, err := cl.Has(ctx, key, ind)
		require.Nil(t, err)
		require.Equal(t, isExist, true)
	}
	res, err := cl.GetL(ctx, key, ind)
	require.Nil(t, err)
	require.Equal(t, res, want)
	value2 := "value_two2"
	err = cl.LPush(ctx, key, value2, ind)
	require.Nil(t, err)
	res, err = cl.GetL(ctx, key, ind)
	want = append(want, value2)
	require.Nil(t, err)
	require.Equal(t, res, want)
	err = cl.DelAll(ctx, key, value, ind)
	require.Nil(t, err)
	want = []string{value2}
	res, err = cl.GetL(ctx, key, ind)
	require.Nil(t, err)
	require.Equal(t, res, want)
	cl.DeleteL(ctx, key, ind)
	require.Nil(t, err)
}
func Test_Client2(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key := "one"
	value := "value_one"
	ind := 0
	want := []string{}
	for i := 0; i < 6; i++ {
		err = cl.LPush(ctx, key, value, ind)
		require.Nil(t, err)
		want = append(want, value)
		isExist, err := cl.Has(ctx, key, ind)
		require.Nil(t, err)
		require.Equal(t, isExist, true)
	}
	res, err := cl.GetL(ctx, key, ind)
	require.Nil(t, err)
	require.Equal(t, res, want)
	err = cl.DelElemL(ctx, key, value, ind)
	require.Nil(t, err)
	want = want[1:]
	res, err = cl.GetL(ctx, key, ind)
	require.Nil(t, err)
	require.Equal(t, res, want)
	err = cl.DeleteL(ctx, key, ind)
	require.Nil(t, err)
	_, err = cl.GetL(ctx, key, ind)
	require.NotNil(t, err)
	isExist, err := cl.Has(ctx, key, ind)
	require.Nil(t, err)
	require.Equal(t, isExist, false)
}
func Test_Client(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	_, err = New(ctx, "localhost:6666", "wrongPassword")
	require.ErrorIs(t, err, ErrInvalidPassword)
	ctx2, _ := context.WithTimeout(ctx, 1*time.Microsecond)
	_, err = New(ctx2, "localhost:6666", "wrongPassword")
	require.ErrorIs(t, err, ErrTimeIsOut)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err = cl.Set(ctx, "foo", "bar", 0)
	require.Nil(t, err)
	val, err := cl.Get(ctx, "foo", 0)
	require.Nil(t, err)
	require.Equal(t, val, "bar", 0)
	cl.Close()

}
func Test_WriteMapResp(t *testing.T) {
	m := map[string]string{
		"foo":  "bar",
		"foo2": "bar2",
	}
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	err = cl.Hello(context.Background(), m)
	require.Nil(t, err)
}

func Test_ADD(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	key := "one"
	require.Nil(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Set(ctx, key, "1", 0)
	require.Nil(t, err)
	val, err := cl.Get(context.Background(), key, 0)
	require.Nil(t, err)
	require.Equal(t, val, "1")
	ctx2, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Add(ctx2, key, 0)
	require.Nil(t, err)
	ctx3, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	val, err = cl.Get(ctx3, key, 0)
	require.Nil(t, err)
	fmt.Println("value")
	fmt.Println(val)
	require.Equal(t, val, "2")
}

func Test_BADVALUE(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key := "one"
	err = cl.Set(context.Background(), key, "badValue", 0)
	require.Nil(t, err)
	err = cl.Add(context.Background(), key, 0)
	require.NotNil(t, err)
	err = cl.AddN(context.Background(), key, "30", 0)
	require.NotNil(t, err)
}

func Test_AddN(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	key := "one"
	value := "2"
	require.Nil(t, err)
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = cl.Set(ctx, key, "1", 0)
	require.Nil(t, err)
	val, err := cl.Get(context.Background(), key, 0)
	require.Nil(t, err)
	require.Equal(t, val, "1")
	err = cl.AddN(ctx, key, value, 0)
	require.Nil(t, err)
	val, err = cl.Get(ctx, key, 0)
	require.Nil(t, err)
	require.Equal(t, val, "3")
}
func Test_ADD_ADDN(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	cl2, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key := "one"
	cl.Set(context.Background(), key, "1", 0)
	start := make(chan struct{})
	wg := sync.WaitGroup{}
	go func() {
		<-start
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl.Add(context.Background(), key, 0)
				require.Nil(t, err)
			}()
		}
	}()
	go func() {
		<-start
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n := rand.Intn(50)
				err := cl2.AddN(context.Background(), key, strconv.Itoa(n), 0)
				require.Nil(t, err)
			}()
		}
	}()
	close(start)
	time.Sleep(1 * time.Millisecond)
	wg.Wait()
	val, err := cl.Get(context.Background(), key, 0)
	require.Nil(t, err)
	fmt.Println(val)
}
func Test_Delete(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key := "one"
	ctx, _ := context.WithTimeout(context.Background(), 500*time.Microsecond)
	err = cl.Set(ctx, key, "1", 0)
	require.Nil(t, err)
	err = cl.Delete(ctx, key, 0)
	require.Nil(t, err)
	_, err = cl.Get(ctx, key, 0)
	require.NotNil(t, err)
}
func Test_DataBaseSupport(t *testing.T) {
	cl, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	cl2, err := New(ctx, "localhost:6666", "")
	require.Nil(t, err)
	key1 := "one"
	val1 := "value_one"
	val2 := "value_two"
	ind1 := 0
	ind2 := 1
	err = cl.Set(context.Background(), key1, val1, ind1)
	require.Nil(t, err)
	err = cl.Set(context.Background(), key1, val2, ind2)
	require.Nil(t, err)
	val, err := cl2.Get(context.Background(), key1, ind1)
	require.Nil(t, err)
	require.Equal(t, val, val1)
	val, err = cl2.Get(context.Background(), key1, ind2)
	require.Nil(t, err)
	require.Equal(t, val, val2)
}
func Test_DataBaseSupport2(t *testing.T) {
	address := "localhost:6666"
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 400; i++ {
		cl, err := New(ctx, address, "")
		require.Nil(t, err)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				value := fmt.Sprintf("value_%v", j)
				key := fmt.Sprintf("myKey_%v", j)
				err := cl.Set(context.Background(), key, value, i)
				if i > 39 {
					require.NotNil(t, err)
				} else {
					require.Nil(t, err)
				}
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 400; i++ {
		cl, err := New(ctx, address, "")
		require.Nil(t, err)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				value := fmt.Sprintf("value_%v", j)
				key := fmt.Sprintf("myKey_%v", j)
				val, err := cl.Get(context.Background(), key, i)
				if i > 39 {
					require.NotNil(t, err)
				} else {
					require.Nil(t, err)
					require.Equal(t, val, value)
				}
			}
		}()
	}
	wg.Wait()
	fmt.Println(time.Since(start))
}
