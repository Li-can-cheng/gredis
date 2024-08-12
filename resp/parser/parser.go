package parser

import (
	"bufio"
	"errors"
	"gredis/interface/resp"
	"gredis/lib/logger"
	"gredis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type Payload struct {
	Data resp.Reply
	Err  error
}

// 读取状态（参数）
type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64 // 用$规定的读取bulk的长度
}

func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Recovered from panic: " + string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for {
		logger.Info("Waiting to read a new line...")
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			logger.Error("Error reading line: " + err.Error())
			if ioErr {
				logger.Error("IO error, closing channel.")
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			ch <- &Payload{Err: err}
			state = readState{}
			continue
		}
		// 不是多行解析模式
		if !state.readingMultiLine {
			if msg[0] == '*' { //*3\r\n
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: err}
					state = readState{}
					continue
				}
				// 如果没有读到多行头部，直接解析单行
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: reply.EmptyMultiBulkReply{}}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' { //*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n

				err := parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: err}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					ch <- &Payload{Data: reply.NullBulkReply{}}
					state = readState{}
					continue
				}

			} else {
				res, err := parseSingleLine(msg)
				ch <- &Payload{Data: res, Err: err}
				state = readState{}
				continue
			}

		} else {
			err := readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg) + err.Error()),
				}
				state = readState{}
				continue
			}
			if state.finished() {
				var res resp.Reply
				switch state.msgType {
				case '*':
					res = reply.MakeMultiBulkReply(state.args)
				case '$':
					res = reply.MakeBulkReply(state.args[0])
				default:

				}
				ch <- &Payload{Data: res, Err: nil}
				state = readState{}
			}
		}
	}
}

// bool代表一些IO错误，error代表解析错误
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 {
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			logger.Error("Failed to read line: " + err.Error())
			return nil, false, err
		}
		logger.Info("Read line: " + string(msg))
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: expected '\\r\\n'" + string(msg))
		}
	} else {
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			logger.Error("Failed to read full bulk data: " + err.Error())
			return nil, true, err
		}
		logger.Info("Read bulk data: " + string(msg))
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: expected '\\r\\n'" + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// *3\r\n
// *3\r\n
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	// 去除消息末尾的 \r\n
	msgStr := strings.TrimSuffix(string(msg), "\r\n")

	// 解析 * 之后的部分
	expectedLine, err := strconv.ParseUint(msgStr[1:], 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg) + " - " + err.Error())
	}

	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)

		// 新增的处理逻辑
		if len(msg) == len(msgStr)+2 && state.expectedArgsCount == 0 {
			return errors.New("protocol error: unexpected empty bulk")
		}

		return nil

	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// +OK -err
func parseSingleLine(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeStandardErrorReply(str[1:])
	case ':':
		num, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(num)
	default:
	}
	return result, nil
}

// $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalve\r\n
// $4\r\nPING\r\n
func readBody(msg []byte, state *readState) error {
	// 检查 msg 长度是否足够
	if len(msg) < 2 {
		return errors.New("protocol error: message too short")
	}

	line := msg[0 : len(msg)-2]
	var err error
	if len(line) > 0 && line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return err
		}
		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
