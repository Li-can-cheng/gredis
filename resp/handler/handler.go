package handler

import (
	"context"
	"gredis/cluster"
	"gredis/config"
	"gredis/database"
	databaseFace "gredis/interface/database"
	"gredis/lib/logger"
	"gredis/lib/sync/atomic"
	"gredis/resp/conn"
	"gredis/resp/parser"
	"gredis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown error\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databaseFace.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseFace.Database
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.NewClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}
	db = database.NewStandaloneDatabase()
	return &RespHandler{
		db: db,
	}
}

func (r *RespHandler) closeClient(client *conn.Conn) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}
func (r *RespHandler) Handle(ctx context.Context, c net.Conn) {
	if r.closing.Get() {
		_ = c.Close()
	}
	client := conn.NewConn(c)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(c)
	logger.Info("connection open" + client.RemoteAddr().String())
	for payload := range ch {
		logger.Info("receive payload")
		//parse error
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("connection close" + client.RemoteAddr().String())
				return
			}
			//protocol error
			errReply := reply.MakeStandardErrorReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("connection close" + client.RemoteAddr().String())
				return
			}
			continue
		}

		if payload.Data == nil {
			continue
		}
		myReply := (payload.Data).(*reply.MultiBulkReply)
		res := r.db.Exec(client, myReply.Args)
		if res == nil {
			_ = client.Write(res.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}

	}
}

func (r *RespHandler) Close() error {
	logger.Info("handler closing")
	r.closing.Set(true)
	r.activeConn.Range(
		func(key, value interface{}) bool {
			client := key.(*conn.Conn)
			_ = (*client).Close()
			return true
		})

	r.db.Close()
	return nil
}
