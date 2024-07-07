package reclogs

import (
	"fmt"
	"testing"

	"github.com/ArtemNovok/simpleRedisCl/interanl/command"
	"github.com/stretchr/testify/require"
)

func Test_WriteLog(t *testing.T) {
	ch := make(chan command.Command)
	st := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-ch:
				switch msg.(type) {
				case command.SetCommand:
					fmt.Println("got set command")
				}
			case <-st:
				fmt.Println("done")
				return
			}
		}
	}()
	r := New("test", ch)
	err := r.WriteLog("SET", 0, []byte("my_key"), []byte("myval"))
	require.Nil(t, err)
	err = r.ReadLog()
	require.Nil(t, err)

}
