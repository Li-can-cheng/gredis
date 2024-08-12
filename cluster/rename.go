package cluster

import (
	"gredis/interface/resp"
	"gredis/resp/reply"
)

// Rename renames a key, the origin and the destination must within the same node
func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeStandardErrorReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.PeerPicker.PickNode(src)
	destPeer := cluster.PeerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeStandardErrorReply("ERR rename must within one slot in cluster mode")
	}
	return cluster.relay(srcPeer, c, args)
}
