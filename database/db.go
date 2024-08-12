package database

import (
	"gredis/data_struct/dict"
	"gredis/interface/database"
	"gredis/interface/resp"
	"gredis/resp/reply"
	"strings"
)

type DB struct {
	index  int
	data   dict.Dict
	addAof func(line CmdLine)
}
type CmdLine = [][]byte

func makeDB() *DB {
	return &DB{
		data: dict.MakeSyncDict(),
		addAof: func(line CmdLine) {

		},
	}
}

type ExecFunc func(db *DB, acmLi [][]byte) resp.Reply

func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	//PING SET SETNX GET GETSET MGET MSET MSETNX DEL
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply("ERR wrong number of arguments for '" + cmdName + "' command")
	}
	fun := cmd.executor

	return fun(db, cmdLine)

}

// set k v 	arity = 3
// exists k1 k2 arity = -2
func validateArity(arity int, cmdArgs CmdLine) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	} else {
		return argNum == -arity
	}
}

// get k
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity := val.(*database.DataEntity)
	return entity, true
}

func (db *DB) SetEntity(key string, entity *database.DataEntity) int {
	return db.data.Set(key, entity)
}

func (db *DB) SetIfExistsEntity(key string, entity *database.DataEntity) int {
	return db.data.SetIfExist(key, entity)
}

func (db *DB) SetIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.SetIfAbsent(key, entity)
}

func (db *DB) Remove(key string) int {
	return db.data.Remove(key)
}

func (db *DB) Removes(keys ...string) (deleted int) {
	for _, key := range keys {
		_, ok := db.data.Get(key)
		if ok {
			db.data.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}
