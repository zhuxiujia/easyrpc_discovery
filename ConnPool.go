package easyrpc_discovery

import "github.com/zhuxiujia/easyrpc"

const ConnError = "connection is shut down"

type ConnPool struct {
	connMap              map[string]*easyrpc.Client
	RpcConnectionFactory RpcConnectionFactory
}

func (it ConnPool) New() ConnPool {
	it.connMap = make(map[string]*easyrpc.Client)
	return it
}

func (it *ConnPool) GetAndPush(serviceName string, addr string) (c *easyrpc.Client, e error) {
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

func (it *ConnPool) Pop(addr string) {
	var conn = it.connMap[addr]
	if conn != nil {
		conn.Close()
		delete(it.connMap, addr)
	}
}
