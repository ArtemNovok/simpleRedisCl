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
	CommnadHello               = "HELLO"
	CommandAdd                 = "ADD"
	CommandAddN                = "ADDN"
	CommandDelete              = "DEL"
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
type AddCommand struct {
	Key []byte
}
type AdddNCommand struct {
	Key []byte
	Val []byte
}
type HelloCommand struct {
	value string
}
type DeleteCommnad struct {
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
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				return SetCommand{
					Key: v.Array()[1].Bytes(),
					Val: v.Array()[2].Bytes(),
				}, nil

			case CommandGet:
				if len(v.Array()) != 2 {
					return nil, ErrUnknownCommandArguments
				}
				return GetCommand{
					Key: v.Array()[1].Bytes(),
				}, nil
			case CommandAdd:
				if len(v.Array()) != 2 {
					return nil, ErrUnknownCommandArguments
				}
				return AddCommand{
					Key: v.Array()[1].Bytes(),
				}, nil
			case CommandAddN:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				return AdddNCommand{
					Key: v.Array()[1].Bytes(),
					Val: v.Array()[2].Bytes(),
				}, nil
			case CommandDelete:
				if len(v.Array()) != 2 {
					return nil, ErrUnknownCommandArguments
				}
				return DeleteCommnad{
					Key: v.Array()[1].Bytes(),
				}, nil
			case CommnadHello:
				return HelloCommand{
					value: v.Array()[1].String(),
				}, nil

			default:
				return nil, ErrUnknownCommand
			}
		}
	}
	return nil, ErrUnknownCommand
}
