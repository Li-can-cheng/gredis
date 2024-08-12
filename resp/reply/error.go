package reply

type UnknownErrReply struct {
}

var unknownErrBytes = []byte("-ERR unknown error\r\n")

func (u UnknownErrReply) Error() string {
	return "unknown error"
}

func (u UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

type ArgNumErrReply struct {
	Cmd string
}

func (r *ArgNumErrReply) Error() string {
	return "wrong number of arguments for '" + r.Cmd + "' command"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

// SyntaxErrReply 语法错误回复
type SyntaxErrReply struct {
}

func (r *SyntaxErrReply) Error() string {
	return "syntax error"
}

var syntaxErrBytes = []byte("-ERR syntax error\r\n")

func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

// WrongTypeErrReply 错误类型回复
type WrongTypeErrReply struct {
}

func (r *WrongTypeErrReply) Error() string {
	return "Operation against a key holding the wrong kind of value"
}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return &WrongTypeErrReply{}
}

// ProtocolErrReply 协议错误回复
type ProtocolErrReply struct {
}

func (r *ProtocolErrReply) Error() string {
	return "Protocol error"
}

var protocolErrBytes = []byte("-ERR Protocol error\r\n")

func (r *ProtocolErrReply) ToBytes() []byte {
	return protocolErrBytes
}

func MakeProtocolErrReply() *ProtocolErrReply {
	return &ProtocolErrReply{}
}
