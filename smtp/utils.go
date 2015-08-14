package smtp

import (
	"fmt"
	"strconv"
	"strings"
)

// Wrap a byte slice paragraph for use in SMTP header
func wrap(sl []byte) []byte {
	length := 0
	for i := 0; i < len(sl); i++ {
		if length > 76 && sl[i] == ' ' {
			sl = append(sl, 0, 0)
			copy(sl[i+2:], sl[i:])
			sl[i] = '\r'
			sl[i+1] = '\n'
			sl[i+2] = '\t'
			i += 2
			length = 0
		}
		if sl[i] == '\n' {
			length = 0
		}
		length++
	}
	return sl
}

func parseCommand(data []byte) (cmd SMTPCommand) {
	line := string(data)

	cmd.Packet = line
	cmd.Fields = strings.Fields(line)

	if len(cmd.Fields) > 0 {
		cmd.Action = strings.ToUpper(cmd.Fields[0])
		if len(cmd.Fields) > 1 {
			cmd.Params = strings.Split(cmd.Fields[1], ":")
		}
	}

	// if cmd.Action == "RCPT" {
	// 	cmd.From = cmd.Params[1]
	// 	cmd.To = cmd.Params[1]
	// }
	//
	// if cmd.Action == "RCPT" {
	// 	cmd.From = cmd.Params[1]
	// 	cmd.To = cmd.Params[1]
	// }

	return
}

func parseCodeLine(line string) (code int, message string, finished bool, err error) {
	if len(line) < 4 || line[3] != ' ' && line[3] != '-' {
		err = fmt.Errorf("Short Response: %s", line)
		return
	}

	finished = line[3] != '-'

	code, err = strconv.Atoi(line[0:3])

	if err != nil {
		return
	}

	if code < 100 {
		err = fmt.Errorf("InValid Response Code: %d for Line: %s", code, line)
		return
	}

	message = line[4:]

	return
}

//Close provides a basic io.WriteCloser write method
func (w *FuncWriter) Close() error {
	w.fx = nil
	return nil
}

//Write provides a basic io.Writer write method
func (w *FuncWriter) Write(b []byte) (int, error) {
	w.fx(b)
	return len(b), nil
}

//NewFuncWriter returns a new function writer instance
func NewFuncWriter(fx func([]byte)) *FuncWriter {
	return &FuncWriter{fx}
}

type (
	//FuncWriter provides a means of creation io.Writer on functions
	FuncWriter struct {
		fx func([]byte)
	}
)
