package easyrpc_discovery

import "github.com/zhuxiujia/easyrpc"

type RpcClient struct {
	Address           string      //远程地址
	Object            interface{} //rpc对象
	Pool              *ConnPool
	LoadBalanceClient *RpcLoadBalanceClient
}

func (it *RpcClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	var e = it.Object.(*easyrpc.Client).Call(serviceMethod, args, reply)
	if e != nil && e.Error() == ConnError {
		it.Close()
	}
	return e
}

func (it *RpcClient) Close() {
	if it.Object != nil {
		it.Object.(*easyrpc.Client).Close()
	}
	if it.LoadBalanceClient != nil {
		it.LoadBalanceClient.Delete(it.Address, nil)
	}
	if it.Pool != nil {
		it.Pool.Pop(it.Address)
	}
}
