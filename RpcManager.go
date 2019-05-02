package easyrpc_discovery

//对于同一个资源的读写必须是原子化的，也就是说，同一时间只能有一个goroutine对共享资源进行读写操作
import (
	"errors"
	"github.com/zhuxiujia/easyrpc"
	"github.com/zhuxiujia/easyrpc/easy_jsonrpc"
	"log"
	"net"
	"strconv"
	"time"
)

var rpcConnectionFactory = RpcConnectionFactory{}
var lastTime = time.Now()

//定义一个服务发现客户端
func EnableDiscoveryClient(consulAddress string, clientName string, client_address string, client_port int, duration time.Duration, config *RpcConfig, serviceBeanArray []RpcServiceBean, registerClient bool) {
	var client = CreateConsulApiClient(consulAddress)
	var serviceId = clientName + ":" + strconv.Itoa(client_port)
	var reg = CreateAgentServiceRegistration(TCP, serviceId, clientName, client_address, client_port)
	var manager = RpcServiceManager{}.New()
	if config != nil {
		manager.RpcConfig = *config
	}
	StartTimer(StartType_Now, Execute_coroutine, duration, func() {
		FetchServiceMap(clientName, &manager, client, deleteClient, clearAllClient)
		if registerClient == false {
			return
		}
		lastTime = time.Now()
		DoRegister(reg, client)
		reg.Tags = []string{clientName + PrintTimeString(" ", lastTime, time.Now(), time.Millisecond)}
	})
	var fullAddress = client_address + strconv.Itoa(client_port)

	for _, v := range serviceBeanArray {
		var remoteServiceName = v.RemoteServiceName
		//Todo create link
		var getClientFunc = func(arg *RpcClient) error {
			return LoadBalance(&manager, arg, fullAddress, remoteServiceName)
		}
		ProxyClient(v, getClientFunc, manager.RpcConfig.RetryTime)
	}
}

//定义一个服务发现服务端
func EnableDiscoveryService(consulAddress string, serviceName string, serviceBeans map[string]interface{}, server_address string, server_port int, duration time.Duration) {
	//轮询注册 服务发现
	var serviceId = serviceName + ":" + strconv.Itoa(server_port)
	var reg = CreateAgentServiceRegistration(TCP, serviceId, serviceName, server_address, server_port)
	var client = CreateConsulApiClient(consulAddress)
	var lastTime = time.Now()
	StartTimer(StartType_Now, Execute_coroutine, duration, func() {
		lastTime = time.Now()
		DoRegister(reg, client)
		reg.Tags = []string{serviceName + PrintTimeString(" ", lastTime, time.Now(), time.Millisecond)}
	})
	//注册Rpc服务
	for _, v := range serviceBeans {
		easyrpc.Register(v)
	}
	if server_address == "localhost" || server_address == "127.0.0.1" {
		server_address = ""
	}
	var tcpUrl = server_address + ":" + strconv.Itoa(server_port)

	l, e := net.Listen("tcp", tcpUrl)
	if e != nil {
		log.Fatalf("net rpc.Listen tcp :0: %v", e)
		panic(e)
	}
	for {
		conn, e := l.Accept()
		if e != nil {
			continue
		}
		go easy_jsonrpc.ServeConn(conn)
	}
}

/**
 * 随机选取服务
 */
func LoadBalance(manager *RpcServiceManager, arg *RpcClient, clientAddr string, remoteService string) error {
	var rpcLoadBalanceClient = manager.ServiceAddressMap[remoteService]

	if arg != nil && arg.Shutdown == true {
		if rpcLoadBalanceClient != nil {
			rpcLoadBalanceClient.Delete(arg.Address, func(client *RpcClient) {
				if client == nil {
					return
				}
				rpcConnectionFactory.Close(client.Object.(*easyrpc.Client))
			})
		}
		arg.Object = nil
	}

	if rpcLoadBalanceClient == nil || len(rpcLoadBalanceClient.RpcClientsMap) == 0 {
		return errors.New("no service '" + remoteService + "' available!")
	}
	var rpcClient = DoBalance(clientAddr, rpcLoadBalanceClient, nil)
	if rpcClient == nil {
		return errors.New("no service '" + remoteService + "' available!")
	}
	if rpcClient.Object == nil {
		rpcClient.Object = createClient(remoteService, rpcClient.Address)
		if rpcClient.Object == nil {
			return errors.New("no service '" + remoteService + "' available!")
		}
	}
	*arg = *rpcClient
	return nil
}

func createClient(serviceName string, address string) interface{} {
	client, err := rpcConnectionFactory.GetConnection(serviceName, address)
	if err != nil {
		log.Println(err)
		return nil
	} else {
		return client
	}
}
func deleteClient(serviceName string, rpcClient *RpcClient) {
	if rpcClient != nil && rpcClient.Object != nil {
		rpcConnectionFactory.Close(rpcClient.Object.(*easyrpc.Client))
	}
}
func clearAllClient(m map[string]*RpcLoadBalanceClient) {
	for _, item := range m {
		if item != nil {
			item.DeleteAll(func(client *RpcClient) {
				rpcConnectionFactory.Close(client.Object.(*easyrpc.Client))
			})
		}
	}
}
