package reply

type PongReply struct {
}

var pongBytes = []byte("+PONG\r\n")

func (r PongReply) ToBytes() []byte {
	return pongBytes
}

// MakePongReply
// 这里是一个工厂方法，用于创建PongReply对象,但每次都是新创建的对象
func MakePongReply() *PongReply {
	return &PongReply{}
}

type OkReply struct {
}

var okBytes = []byte("+OK\r\n")

func (r OkReply) ToBytes() []byte {
	return okBytes
}

// 预先创建好的OkReply对象，避免重复创建，节约内存
var theOkReply = new(OkReply)

// MakeOkReply
// 这里是一个工厂方法，用于创建OkReply对象
func MakeOkReply() *OkReply {
	return theOkReply
}

// NullBulkReply 空的块回复
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")

func (r NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// EmptyMultiBulkReply 空的多块回复
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkBytes = []byte("*0\r\n")

func (r EmptyMultiBulkReply) ToBytes() []byte {
	//TODO implement me
	panic("implement me")
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

type NoReply struct {
}

var noBytes = []byte("")

func (r NoReply) ToBytes() []byte {
	return noBytes
}

var theNoReply = new(NoReply)

func MakeNoReply() *NoReply {
	return theNoReply
}
