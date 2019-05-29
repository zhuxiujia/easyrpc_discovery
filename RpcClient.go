package easyrpc_discovery

import "github.com/zhuxiujia/easyrpc"

type RpcClient struct {
	Address  string      //远程地址
	Object   interface{} //rpc对象
	Shutdown bool        // server has told us to stop
	Pool     *ConnPool
}

func (it *RpcClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	var e = it.Object.(*easyrpc.Client).Call(serviceMethod, args, reply)
	if e != nil && e.Error() == ConnError {
		if it.Pool != nil {
			it.Pool.Pop(it.Address)
		}
	}
	return e
}
