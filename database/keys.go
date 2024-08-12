package database

import (
	"gredis/interface/resp"
	"gredis/lib/utils"
	"gredis/lib/wildcard"
	"gredis/resp/reply"
)

// del k1 k2 k3
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

func init() {
	RegisterCommand("del", execDel, -2)
	RegisterCommand("exists", execExists, -2)
	RegisterCommand("flushdb", execFlushDB, -1)
	RegisterCommand("type", execType, 1)
	RegisterCommand("rename", execRename, 3)
	RegisterCommand("renamenx", execRenamenx, 3)
	RegisterCommand("keys", execKeys, 2)
}

// exists k1 k2 k3 ...
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, ok := db.GetEntity(key)
		if ok {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// keys
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	res := make([][]byte, 0)
	db.data.ForEach(func(key string, value interface{}) bool {
		if pattern.IsMatch(key) {
			res = append(res, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(res)
}

// flushdb
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOkReply()
}

// type
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	//TODO: add more types
	return &reply.UnknownErrReply{}
}

// rename
func execRename(db *DB, args [][]byte) resp.Reply {
	oldKey := string(args[0])
	newKey := string(args[1])
	entity, ok := db.GetEntity(oldKey)
	if !ok {
		return reply.MakeStandardErrorReply("no such key")
	}
	db.SetEntity(newKey, entity)
	db.Remove(oldKey)
	db.addAof(utils.ToCmdLine2("rename", args...))

	return reply.MakeOkReply()
}

// renamenx k1 k2
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	oldKey := string(args[0])
	newKey := string(args[1])
	_, exists := db.GetEntity(newKey)
	if exists {
		return reply.MakeIntReply(0)
	}
	entity, ok := db.GetEntity(oldKey)
	if !ok {
		return reply.MakeStandardErrorReply("no such key")
	}
	db.SetEntity(newKey, entity)
	db.Remove(oldKey)
	db.addAof(utils.ToCmdLine2("renamenx", args...))

	return reply.MakeIntReply(1)
}
