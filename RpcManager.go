package easyrpc_discovery

//对于同一个资源的读写必须是原子化的，也就是说，同一时间只能有一个goroutine对共享资源进行读写操作
import (
	"errors"
	"github.com/zhuxiujia/easyrpc"
	"github.com/zhuxiujia/easyrpc/easy_jsonrpc"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

var rpcConnectionFactory = RpcConnectionFactory{}

// Register Service Register
type Register interface {
	InitRegister(
		serviceName string,
		address string,
		port int,
		duration time.Duration)

	DoRegister()
}

// ServiceFetcher Service List fetcher
type ServiceFetcher interface {
	InitServiceFetcher(manager *RpcServiceManager,
		clearFunc func(m map[string]*RpcLoadBalanceClient),
		pool *ConnPool)

	DoFetch()
}

//定义一个服务发现客户端
func EnableDiscoveryClient(
	balanceType *LoadBalanceType,
	clientName string,
	clientAddress string,
	clientPort int,
	duration time.Duration,
	config *RpcConfig,
	serviceBeanArray []RpcServiceBean, reg Register, fetch ServiceFetcher) {
	var manager = RpcServiceManager{}.New()
	if config != nil {
		manager.RpcConfig = *config
	}
	var pool = ConnPool{}.New()

	if reg != nil {
		reg.InitRegister(clientName, clientAddress, clientPort, duration)
	}
	if fetch != nil {
		fetch.InitServiceFetcher(&manager, clearAllClient, &pool)
	}
	var fullAddress = clientAddress + strconv.Itoa(clientPort)
	StartTimer(StartType_Now, Execute_coroutine, duration, func() {
		//FetchServiceMap(clientName, &manager, client, clearAllClient, &pool)
		//if registerClient == false {
		//	return
		//}
		//DoRegister(reg, client)

		if fetch != nil {
			fetch.DoFetch()
		}
		if reg != nil {
			reg.DoRegister()
		}
	})

	var getClientFunc = func(RemoteServiceName string) (c *RpcClient, e error) {
		return LoadBalance(&manager, fullAddress, RemoteServiceName, balanceType)
	}

	for _, v := range serviceBeanArray {
		ProxyClient(v, getClientFunc)
	}
}

//定义一个服务发现服务端
func EnableDiscoveryService(
	serviceBeans map[string]interface{},
	server_address string,
	server_port int,
	duration time.Duration,
	deferFunc func(recover interface{}) string,
	newRegisterFunc func() Register,
) {
	//注册Rpc服务
	var funcs = []func(){}
	for _, v := range serviceBeans {
		serviceName := reflect.TypeOf(v).Elem().Name()
		//轮询注册 服务发现
		var reg = newRegisterFunc()
		reg.InitRegister(serviceName, server_address, server_port, duration)
		funcs = append(funcs, func() {
			reg.DoRegister()
		})
		easyrpc.RegisterDefer(v, deferFunc)
	}
	StartTimer(StartType_Now, Execute_coroutine, duration, func() {
		for _, item := range funcs {
			item()
		}
	})
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
func LoadBalance(manager *RpcServiceManager, clientAddr string, remoteService string, balanceType *LoadBalanceType) (client *RpcClient, e error) {
	var rpcLoadBalanceClient = manager.ServiceAddressMap[remoteService]

	if rpcLoadBalanceClient == nil || len(rpcLoadBalanceClient.RpcClientsMap) == 0 {
		return nil, errors.New("no service '" + remoteService + "' available!")
	}
	var rpcClient = DoBalance(clientAddr, rpcLoadBalanceClient, balanceType)
	if rpcClient == nil {
		return nil, errors.New("no service '" + remoteService + "' available!")
	}
	return rpcClient, nil
}

func deleteClient(rpcClient *RpcClient) {
	if rpcClient != nil {
		rpcClient.Close()
	}
}
func clearAllClient(m map[string]*RpcLoadBalanceClient) {
	for _, item := range m {
		if item != nil {
			item.DeleteAll(func(client *RpcClient) {
				if client != nil {
					client.Close()
				}
			})
		}
	}
}
