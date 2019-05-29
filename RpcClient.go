package easyrpc_discovery

type RpcClient struct {
	Address           string //远程地址
	Pool              *ConnPool
	LoadBalanceClient *RpcLoadBalanceClient
	ShutDown          bool
}

func (it *RpcClient) Call(serviceName string, serviceMethod string, args interface{}, reply interface{}) error {
	var c, e = it.Pool.GetAndPush(serviceName, it.Address)
	if e != nil {
		return e
	}
	e = c.Call(serviceMethod, args, reply)
	if e != nil && e.Error() == ConnError {
		it.Close()
	}
	return e
}

func (it *RpcClient) Close() {
	if it.LoadBalanceClient != nil {
		it.LoadBalanceClient.Delete(it.Address, nil)
	}
	if it.Pool != nil {
		it.Pool.Pop(it.Address)
	}
	it.ShutDown = true
}
