package database

import "gredis/interface/resp"

type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(client resp.Connection)
}

type DataEntity struct {
	Data any
}
