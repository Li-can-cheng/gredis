package cluster

import (
	"gredis/interface/resp"
	"gredis/resp/reply"
)

func flushdb(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdAndArgs)
	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrorReply(r) {

			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOkReply()
	}
	return reply.MakeStandardErrorReply(errReply.Error())
}
