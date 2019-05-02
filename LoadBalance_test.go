package easyrpc_discovery

import (
	"testing"
	"time"
)

func TestDoBalance(t *testing.T) {
	var clients = make([]*RpcClient, 0)
	clients = append(clients, &RpcClient{
		Address: "127.0.0.1:1234",
	})
	clients = append(clients, &RpcClient{
		Address: "127.0.0.1:1230",
	})
	clients = append(clients, &RpcClient{
		Address: "127.0.0.1:1231",
	})
	clients = append(clients, &RpcClient{
		Address: "127.0.0.1:1232",
	})
	var c = RpcLoadBalanceClient{}
	c.SetRpcClients(clients)
	var lt = LoadBalance_Random

	var total = 1000000

	defer CountMethodUseTime(time.Now(), "TestDoBalance", time.Millisecond)

	for i := 0; i < total; i++ {
		DoBalance("127.0.0.1", &c, &lt)
		//fmt.Println(DoBalance("127.0.0.1",&c,&lt))//fmt比较因为使用了反射，性能较差，测试性能建议注释它
	}
}
