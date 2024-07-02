package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/tidwall/resp"
)

var (
	CommandSet                 = "SET"
	ErrUnknownCommand          = errors.New("unknown command")
	ErrUnknownCommandArguments = errors.New("unknown command arguments")
)

type Command interface {
	// TODO
}

type SetCommand struct {
	key, val string
}

func ParseCommand(rawMsg string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(rawMsg))
	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		fmt.Printf("Read %s\n", v.Type())
		if v.Type() == resp.Array {
			switch v.Array()[0].String() {
			case CommandSet:
				if len(v.Array()) > 3 {
					return nil, ErrUnknownCommandArguments
				}
				return SetCommand{
					key: v.Array()[1].String(),
					val: v.Array()[2].String(),
				}, nil
			default:
				return nil, ErrUnknownCommand

			}
		}
	}
	return nil, ErrUnknownCommand
}