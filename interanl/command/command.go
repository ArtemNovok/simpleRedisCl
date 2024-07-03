package command

import (
	"bytes"
	"errors"
	"io"

	"github.com/tidwall/resp"
)

var (
	CommandSet                 = "SET"
	CommandGet                 = "GET"
	ErrUnknownCommand          = errors.New("unknown command")
	ErrUnknownCommandArguments = errors.New("unknown command arguments")
)

type Command interface {
	// TODO
}

type SetCommand struct {
	Key, Val []byte
}
type GetCommand struct {
	Key []byte
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
		if v.Type() == resp.Array {
			switch v.Array()[0].String() {
			case CommandSet:
				if len(v.Array()) > 3 {
					return nil, ErrUnknownCommandArguments
				}
				return SetCommand{
					Key: v.Array()[1].Bytes(),
					Val: v.Array()[2].Bytes(),
				}, nil

			case CommandGet:
				if len(v.Array()) > 2 {
					return nil, ErrUnknownCommandArguments
				}
				return GetCommand{
					Key: v.Array()[1].Bytes(),
				}, nil
			default:
				return nil, ErrUnknownCommand
			}
		}
	}
	return nil, ErrUnknownCommand
}
