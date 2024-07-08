package reclogs

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ArtemNovok/simpleRedisCl/interanl/command"
	"github.com/stretchr/testify/require"
)

func Test_WriteLog(t *testing.T) {
	ch := make(chan command.Command)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg := <-ch
			switch msg.(type) {
			case command.SetCommand:
				fmt.Println("got set command")
			case command.AddCommand:
				fmt.Println("got add command")
			case command.StopCommnad:
				fmt.Println("done")
				return
			}

		}
	}()
	r := New("test", ch)
	err := r.WriteLog("SET", 0, []byte("my_key"), []byte("myval"))
	require.Nil(t, err)
	err = r.WriteLog("ADD", 0, []byte("my_key"))
	require.Nil(t, err)
	err = r.ReadLog()
	require.Nil(t, err)
	wg.Wait()
}
