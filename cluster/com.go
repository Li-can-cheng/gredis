package cluster

import (
	"context"
	"errors"
	"gredis/interface/resp"
	"gredis/lib/client"
	"gredis/lib/utils"
	"gredis/resp/reply"
	"strconv"
)

func (c *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	objectPool, ok := c.peerConnection[peer]
	if !ok {
		return nil, errors.New("peer not found")
	}
	object, err := objectPool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	c2, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("object is not a client")
	}
	return c2, nil
}
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}

// relay relays command to peer
// select db by c.GetDBIndex()
// cannot call Prepare, Commit, execRollback of self node
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self {
		// to self db
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	return peerClient.Send(args)
}

// broadcast broadcasts command to all node in cluster
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	result := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		r := cluster.relay(node, c, args)
		result[node] = r
	}
	return result
}
