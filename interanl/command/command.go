package command

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/tidwall/resp"
)

var (
	CommandSet                 = "SET"
	CommandGet                 = "GET"
	CommandHello               = "HELLO"
	CommandAdd                 = "ADD"
	CommandAddN                = "ADDN"
	CommandDelete              = "DEL"
	CommandLPush               = "LPUSH"
	CommandGetL                = "GETL"
	CommandHas                 = "HAS"
	CommandDeleteL             = "DELL"
	CommandDelElemL            = "DELELEML"
	CommandDelAll              = "DELALL"
	CommandStop                = "STOP"
	ErrUnknownCommand          = errors.New("unknown command")
	ErrUnknownCommandArguments = errors.New("unknown command arguments")
	ErrInvalidIndexValue       = errors.New("invalid index value")
)

type Command interface {
	// TODO
}
type StopCommand struct {
}
type DelAllCommand struct {
	Key, Val []byte
	Index    int
}
type DeleteLCommand struct {
	Key   []byte
	Index int
}
type DelElemLCommand struct {
	Key, Val []byte
	Index    int
}
type LPushCommand struct {
	Key, Val []byte
	Index    int
}
type SetCommand struct {
	Key, Val []byte
	Index    int
}
type GetLCommand struct {
	Key   []byte
	Index int
}
type HasCommand struct {
	Key   []byte
	Index int
}
type GetCommand struct {
	Key   []byte
	Index int
}
type AddCommand struct {
	Key   []byte
	Index int
}
type AddNCommand struct {
	Key   []byte
	Val   []byte
	Index int
}
type HelloCommand struct {
	value string
	Index int
}
type DeleteCommand struct {
	Key   []byte
	Index int
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
			case CommandDelAll:
				if len(v.Array()) != 4 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[3].String())
				if err != nil {
					return nil, err
				}
				return DelAllCommand{
					Key:   v.Array()[1].Bytes(),
					Val:   v.Array()[2].Bytes(),
					Index: ind,
				}, nil
			case CommandDeleteL:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return DeleteLCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandDelElemL:
				if len(v.Array()) != 4 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[3].String())
				if err != nil {
					return nil, err
				}
				return DelElemLCommand{
					Key:   v.Array()[1].Bytes(),
					Val:   v.Array()[2].Bytes(),
					Index: ind,
				}, nil
			case CommandLPush:
				if len(v.Array()) != 4 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[3].String())
				if err != nil {
					return nil, err
				}
				return LPushCommand{
					Key:   v.Array()[1].Bytes(),
					Val:   v.Array()[2].Bytes(),
					Index: ind,
				}, nil
			case CommandHas:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return HasCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandGetL:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return GetLCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandSet:
				if len(v.Array()) != 4 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[3].String())
				if err != nil {
					return nil, err
				}
				return SetCommand{
					Key:   v.Array()[1].Bytes(),
					Val:   v.Array()[2].Bytes(),
					Index: ind,
				}, nil

			case CommandGet:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return GetCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandAdd:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return AddCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandAddN:
				if len(v.Array()) != 4 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[3].String())
				if err != nil {
					return nil, err
				}
				return AddNCommand{
					Key:   v.Array()[1].Bytes(),
					Val:   v.Array()[2].Bytes(),
					Index: ind,
				}, nil
			case CommandDelete:
				if len(v.Array()) != 3 {
					return nil, ErrUnknownCommandArguments
				}
				ind, err := strconv.Atoi(v.Array()[2].String())
				if err != nil {
					return nil, err
				}
				return DeleteCommand{
					Key:   v.Array()[1].Bytes(),
					Index: ind,
				}, nil
			case CommandHello:
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
