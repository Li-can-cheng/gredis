package reply

import (
	"bytes"
	"gredis/interface/resp"
	"strconv"
)

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

var (
	nullBulkReplyBytes = []byte("$-1\r\n")
	CRLF               = "\r\n"
)

type BulkReply struct {
	Arg []byte
}

func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

type MultiBulkReply struct {
	Args [][]byte
}

func (m *MultiBulkReply) ToBytes() []byte {
	argLen := len(m.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(argLen) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

type StatusReply struct {
	Status string
}

func (s StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

type IntReply struct {
	Code int64
}

func (i IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

func MakeIntReply(i int64) *IntReply {
	return &IntReply{Code: i}
}

type StandardErrorReply struct {
	Err string
}

func (s StandardErrorReply) Error() string {
	return s.Err
}

func (s StandardErrorReply) ToBytes() []byte {
	return []byte("-" + s.Err + CRLF)
}

func MakeStandardErrorReply(err string) *StandardErrorReply {
	return &StandardErrorReply{Err: err}
}

func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
