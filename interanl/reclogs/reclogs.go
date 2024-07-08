package reclogs

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ArtemNovok/simpleRedisCl/interanl/command"
)

type RecoveryLogger struct {
	mu       sync.Mutex
	FileName string
	recData  chan command.Command
}

func New(filename string, ch chan command.Command) *RecoveryLogger {
	return &RecoveryLogger{
		FileName: filename,
		recData:  ch,
	}
}
func (r *RecoveryLogger) WriteLog(operation string, ind int, args ...[]byte) error {
	const op = "reclogs.WriteLog"
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := os.OpenFile(r.FileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	defer f.Close()
	_ = f
	var log string
	index := strconv.Itoa(ind)
	log += fmt.Sprintf("%s#%s#", operation, index)
	for _, val := range args {
		log += fmt.Sprintf("%s#", string(val))
	}
	log += "\n"
	_, err = f.Write([]byte(log))
	if err != nil {
		return err
	}
	return nil
}
func (r *RecoveryLogger) ReadLog() error {
	const op = "reclogs.ReadLog"
	f, err := os.OpenFile(r.FileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	defer f.Close()
	scaner := bufio.NewScanner(f)
	for scaner.Scan() {
		attrs := strings.Split(scaner.Text(), "#")
		if len(attrs) < 4 {
			break
		}
		attrs = attrs[:len(attrs)-1]
		cmd, err := parseCommand(attrs)
		if err != nil {
			return err
		}
		r.recData <- cmd
	}
	if err := scaner.Err(); err != nil {
		return err
	}
	r.recData <- command.StopCommand{}
	return nil
}

func parseCommand(attrs []string) (command.Command, error) {
	switch attrs[0] {
	case command.CommandSet:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.SetCommand{
			Key:   []byte(attrs[2]),
			Val:   []byte(attrs[3]),
			Index: ind,
		}, nil
	case command.CommandAdd:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.AddCommand{
			Key:   []byte(attrs[2]),
			Index: ind,
		}, nil
	case command.CommandAddN:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.AddNCommand{
			Key:   []byte(attrs[2]),
			Val:   []byte(attrs[3]),
			Index: ind,
		}, nil
	case command.CommandDelete:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.DeleteCommand{
			Key:   []byte(attrs[2]),
			Index: ind,
		}, nil
	case command.CommandLPush:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.LPushCommand{
			Key:   []byte(attrs[2]),
			Val:   []byte(attrs[3]),
			Index: ind,
		}, nil
	case command.CommandDelElemL:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.DelElemLCommand{
			Key:   []byte(attrs[2]),
			Val:   []byte(attrs[3]),
			Index: ind,
		}, nil
	case command.CommandDeleteL:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.DeleteLCommand{
			Key:   []byte(attrs[2]),
			Index: ind,
		}, nil
	case command.CommandDelAll:
		ind, err := strconv.Atoi(attrs[1])
		if err != nil {
			return nil, command.ErrInvalidIndexValue
		}
		return command.DelAllCommand{
			Key:   []byte(attrs[2]),
			Val:   []byte(attrs[3]),
			Index: ind,
		}, nil
	default:
		return nil, command.ErrUnknownCommand
	}
}
