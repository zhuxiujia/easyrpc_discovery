package easyrpc_discovery

type RpcClient struct {
	Address  string      //远程地址
	Object   interface{} //rpc对象
	Shutdown bool        // server has told us to stop
}
