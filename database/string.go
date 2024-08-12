package database

import (
	"gredis/interface/database"
	"gredis/interface/resp"
	"gredis/lib/utils"
	"gredis/resp/reply"
)

// get
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes)

}

// set
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	db.SetEntity(key, &database.DataEntity{Data: value})
	db.addAof(utils.ToCmdLine2("set", args...))

	return reply.MakeOkReply()
}

// setnx
func execSetnx(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	exists := db.SetIfAbsent(key, &database.DataEntity{Data: value})
	db.addAof(utils.ToCmdLine2("setnx", args...))

	return reply.MakeIntReply(int64(exists))
}

// getset
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity, ok := db.GetEntity(key)
	if !ok {
		db.SetEntity(key, &database.DataEntity{Data: value})
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	db.SetEntity(key, &database.DataEntity{Data: value})
	db.addAof(utils.ToCmdLine2("getset", args...))
	return reply.MakeBulkReply(bytes)
}

// strlen
func execStrlen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return reply.MakeIntReply(0)
	}
	bytes := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(bytes)))
}

func init() {
	RegisterCommand("get", execGet, 2)
	RegisterCommand("set", execSet, 3)
	RegisterCommand("setnx", execSetnx, 3)
	RegisterCommand("getset", execGetSet, 3)
	RegisterCommand("strlen", execStrlen, 2)
}
