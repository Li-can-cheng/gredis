package database

import (
	"gredis/interface/resp"
	"gredis/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// 关键字
func init() {
	RegisterCommand("ping", Ping, 1)
}
