package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ArtemNovok/simpleRedisCl/interanl/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ServerAndClients(t *testing.T) {
	logger := setUpLogger()
	s := SetUpServer(logger)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New("localhost:6666")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key2_%v", i)
			val := fmt.Sprintf("val2_%v", i)
			go func() {
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
		go func() {
			err := cl.Set(context.Background(), key, val)
			require.Nil(t, err)
			val2, err := cl.Get(context.Background(), key)
			require.Nil(t, err)
			require.Equal(t, val, val2)
		}()
	}
	time.Sleep(2 * time.Second)
	s.ShowData()

}

func Test_TwoClientWriteOneValue(t *testing.T) {
	logger := setUpLogger()
	s := SetUpServer(logger)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New("localhost:6666")
	if err != nil {
		log.Fatal(err)
	}
	cl2, err := client.New("localhost:6666")
	if err != nil {
		log.Fatal(err)
	}
	startChan := make(chan struct{})
	go func() {
		<-startChan
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val_%v", i)
			go func() {
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
			go func() {
				err := cl2.Set(context.Background(), key, val)
				assert.Nil(t, err)
			}()
		}
	}()
	time.Sleep(200 * time.Millisecond)
	close(startChan)
	time.Sleep(2 * time.Second)
	s.ShowData()
}
func Test_TwoClientWritesAndReadOneValue(t *testing.T) {
	logger := setUpLogger()
	s := SetUpServer(logger)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl, err := client.New("localhost:6666")
	if err != nil {
		log.Fatal(err)
	}
	cl2, err := client.New("localhost:6666")
	if err != nil {
		log.Fatal(err)
	}
	startChan := make(chan struct{})
	go func() {
		<-startChan
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("key_%v", i)
			val := fmt.Sprintf("val_%v", i)
			go func() {
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
			go func() {
				err := cl2.Set(context.Background(), key, val)
				assert.Nil(t, err)
				_, err = cl.Get(context.Background(), key)
				assert.Nil(t, err)
			}()
		}
	}()
	time.Sleep(200 * time.Millisecond)
	close(startChan)
	time.Sleep(2 * time.Second)
	s.ShowData()

}
func setUpLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}

func SetUpServer(logger *slog.Logger) *Server {
	cfg := Config{
		Log: logger,
	}
	return NewServer(cfg)
}
