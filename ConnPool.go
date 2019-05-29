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

func (it *ConnPool) Get(serviceName string, addr string) (c *RpcClient, e error) {
	c = it.clientMap[addr]
	if c == nil {
		conn, e := it.createClient(serviceName, addr)
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

func (it *ConnPool) GetCoon(addr string) *easyrpc.Client {
	return it.connMap[addr]
}

func (it ConnPool) createClient(serviceName string, addr string) (c *easyrpc.Client, e error) {
	var conn = it.GetCoon(addr)
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
