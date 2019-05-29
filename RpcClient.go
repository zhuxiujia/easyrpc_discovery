package easyrpc_discovery

type RpcClient struct {
	Address           string //远程地址
	Pool              *ConnPool
	LoadBalanceClient *RpcLoadBalanceClient
	Retry             int //重试次数
	Shutdown          bool
}

func (it RpcClient) New(addr string, pool *ConnPool, load *RpcLoadBalanceClient, retry int) RpcClient {
	it.Address = addr
	it.Pool = pool
	it.LoadBalanceClient = load
	it.Retry = retry
	return it
}

func (it *RpcClient) Call(serviceName string, serviceAndMethod string, args interface{}, reply interface{}) (e error) {
	if it.Retry == 0 {
		it.Retry = 1
	}
	if it.Retry > 0 {
		for i := 0; i < it.Retry; i++ {
			var c, e = it.Pool.GetAndPush(serviceName, it.Address)
			if e != nil {
				return e
			}
			e = c.Call(serviceAndMethod, args, reply)
			if e != nil && e.Error() == ConnError {
				it.Close()
			}
			if i+1 == it.Retry {
				return e
			}
			if e == nil {
				return nil
			}
		}
	}
	return e
}

func (it *RpcClient) Close() {
	if it.Shutdown {
		return
	}
	if it.LoadBalanceClient != nil {
		it.LoadBalanceClient.Delete(it.Address, nil)
	}
	if it.Pool != nil {
		it.Pool.Pop(it.Address)
	}
	it.Shutdown = true
}
