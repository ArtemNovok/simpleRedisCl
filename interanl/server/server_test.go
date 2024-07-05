package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ArtemNovok/simpleRedisCl/interanl/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ServerAndClients(t *testing.T) {
	logger := setUpLogger()
	addr := ":5555"
	s := SetUpServer(logger, addr)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	go func() {
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key2_%v", i)
			val := fmt.Sprintf("val2_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl.Set(context.Background(), key, val)
				assert.Nil(t, err)
				val2, err := cl.Get(context.Background(), key)
				assert.Nil(t, err)
				assert.Equal(t, val, val2)
			}()
		}
	}()

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%v", i)
		val := fmt.Sprintf("val_%v", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cl.Set(context.Background(), key, val)
			require.Nil(t, err)
			val2, err := cl.Get(context.Background(), key)
			require.Nil(t, err)
			require.Equal(t, val, val2)
		}()
	}
	wg.Wait()
	s.ShowData()
}

func Test_TwoClientWriteOneValue(t *testing.T) {
	wg2 := sync.WaitGroup{}
	logger := setUpLogger()
	addr := ":8888"
	s := SetUpServer(logger, addr)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl2, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	startChan := make(chan struct{})
	go func() {
		<-startChan
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val_%v", i)
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				err := cl.Set(context.Background(), key, val)
				assert.Nil(t, err)
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val2_%v", i)
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				err := cl2.Set(context.Background(), key, val)
				assert.Nil(t, err)
			}()
		}
	}()
	time.Sleep(200 * time.Millisecond)
	close(startChan)
	wg2.Wait()
	time.Sleep(30 * time.Millisecond)
	cl.Close()
	time.Sleep(30 * time.Millisecond)
	require.Equal(t, len(s.peers), 1)
	s.ShowData()
}
func Test_TwoClientWritesAndReadOneValue(t *testing.T) {
	wg := sync.WaitGroup{}
	logger := setUpLogger()
	addr := ":3333"
	s := SetUpServer(logger, addr)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl2, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	startChan := make(chan struct{})
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl.Set(context.Background(), key, val)
				assert.Nil(t, err)
				_, err = cl.Get(context.Background(), key)
				assert.Nil(t, err)
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val2_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl2.Set(context.Background(), key, val)
				assert.Nil(t, err)
				_, err = cl.Get(context.Background(), key)
				assert.Nil(t, err)
			}()
		}
	}()
	time.Sleep(200 * time.Millisecond)
	close(startChan)
	wg.Wait()
	s.ShowData()
}

func Test_FiveClient(t *testing.T) {
	wg := sync.WaitGroup{}
	logger := setUpLogger()
	addr := ":4444"
	s := SetUpServer(logger, addr)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl2, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl3, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl4, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	cl5, err := client.New(fmt.Sprintf("localhost%s", addr))
	if err != nil {
		log.Fatal(err)
	}
	startChan := make(chan struct{})
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl.Set(context.Background(), key, val)
				require.Nil(t, err)
				n := rand.Intn(3)
				go func() {
					time.Sleep(time.Duration(n*100) * time.Millisecond)
					_, err = cl.Get(context.Background(), key)
					require.Nil(t, err)
				}()
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val2_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl2.Set(context.Background(), key, val)
				require.Nil(t, err)
				n := rand.Intn(3)
				go func() {
					time.Sleep(time.Duration(n*100) * time.Millisecond)
					_, err = cl2.Get(context.Background(), key)
					require.Nil(t, err)
				}()
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val3_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl3.Set(context.Background(), key, val)
				require.Nil(t, err)
				n := rand.Intn(3)
				go func() {
					time.Sleep(time.Duration(n*100) * time.Millisecond)
					_, err = cl3.Get(context.Background(), key)
					require.Nil(t, err)
				}()
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val4_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl4.Set(context.Background(), key, val)
				require.Nil(t, err)
				n := rand.Intn(3)
				go func() {
					time.Sleep(time.Duration(n*100) * time.Millisecond)
					_, err = cl4.Get(context.Background(), key)
					require.Nil(t, err)
				}()
			}()
		}
	}()
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val5_%v", i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cl5.Set(context.Background(), key, val)
				require.Nil(t, err)
				n := rand.Intn(3)
				go func() {
					time.Sleep(time.Duration(n*100) * time.Millisecond)
					_, err = cl5.Get(context.Background(), key)
					require.Nil(t, err)
				}()
			}()
		}
	}()
	close(startChan)
	wg.Wait()
	time.Sleep(600 * time.Millisecond)
	s.ShowData()
}

func setUpLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}

func SetUpServer(logger *slog.Logger, addr string) *Server {
	cfg := Config{
		Log:        logger,
		ListenAddr: addr,
	}
	return NewServer(cfg)
}
