package easyrpc_discovery

import "github.com/zhuxiujia/easyrpc"

const ConnError = "connection is shut down"

type ConnPool struct {
	connMap   map[string]*easyrpc.Client
	clientMap map[string]*RpcClient
	RpcConnectionFactory
}

func (it ConnPool) New() ConnPool {
	it.connMap = make(map[string]*easyrpc.Client)
	it.clientMap = make(map[string]*RpcClient)
	return it
}

func (it *ConnPool) GetAndPush(serviceName string, addr string) (c *RpcClient, e error) {
	c = it.clientMap[addr]
	if c == nil {
		conn, e := it.getAndPushConn(serviceName, addr)
		if e != nil {
			return c, e
		}
		c = &RpcClient{
			Address:  addr,
			Object:   conn,
			Shutdown: false,
		}
		it.clientMap[addr] = c
	}
	return c, e
}

func (it *ConnPool) Pop(addr string) {
	var conn = it.connMap[addr]
	if conn != nil {
		conn.Close()
	}
	var client = it.clientMap[addr]
	if client != nil {
		client.Shutdown = true
	}
}

func (it ConnPool) getAndPushConn(serviceName string, addr string) (c *easyrpc.Client, e error) {
	var conn = it.connMap[addr]
	if conn == nil {
		client, e := it.RpcConnectionFactory.GetConnection(serviceName, addr)
		if e != nil {
			return c, e
		}
		conn = client
		it.connMap[addr] = conn
	}
	return conn, e
}
