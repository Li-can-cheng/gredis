package database

import (
	"gredis/aof"
	"gredis/config"
	"gredis/interface/resp"
	"gredis/lib/logger"
	"gredis/resp/reply"
	"strconv"
	"strings"
)

type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

func NewStandaloneDatabase() *StandaloneDatabase {
	database := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := makeDB()
		db.index = i
		database.dbSet[i] = db

	}

	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAOFHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			// copy db to
			cdb := db
			cdb.addAof = func(line CmdLine) {
				database.aofHandler.AddAof(cdb.index, line)
			}
		}
	}

	return database
}

func (database *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	switch cmdName {
	case "select":
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("ERR wrong number of arguments for '" + cmdName + "' command")
		}
		return execSelect(client, database, args[1:])
	}
	dbIndex := client.GetDBIndex()
	return database.dbSet[dbIndex].Exec(client, args)

}

func (database *StandaloneDatabase) Close() {
}

func (database *StandaloneDatabase) AfterClientClose(client resp.Connection) {
}

func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dbSet) {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
