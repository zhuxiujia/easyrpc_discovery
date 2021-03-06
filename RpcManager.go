package easyrpc_discovery

//对于同一个资源的读写必须是原子化的，也就是说，同一时间只能有一个goroutine对共享资源进行读写操作
import (
	"errors"
	"fmt"
	"github.com/zhuxiujia/easyrpc"
	"github.com/zhuxiujia/easyrpc/easy_jsonrpc"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

var rpcConnectionFactory = RpcConnectionFactory{}

//定义一个服务发现客户端
func EnableDiscoveryClient(balanceType *LoadBalanceType, consulAddress string, clientName string, client_address string, client_port int, duration time.Duration, config *RpcConfig, serviceBeanArray []RpcServiceBean, registerClient bool) {
	var client = CreateConsulApiClient(consulAddress)
	var serviceId = clientName + ":" + strconv.Itoa(client_port)
	var reg = CreateAgentServiceRegistration(TCP, serviceId, clientName, client_address, client_port, fmt.Sprint(duration.Seconds()))
	var manager = RpcServiceManager{}.New()
	if config != nil {
		manager.RpcConfig = *config
	}
	var pool = ConnPool{}.New()
	var fullAddress = client_address + strconv.Itoa(client_port)
	StartTimer(StartType_Now, Execute_coroutine, duration, func() {
		FetchServiceMap(clientName, &manager, client, clearAllClient, &pool)
		if registerClient == false {
			return
		}
		DoRegister(reg, client)
	})

	var getClientFunc = func(RemoteServiceName string) (c *RpcClient, e error) {
		return LoadBalance(&manager, fullAddress, RemoteServiceName, balanceType)
	}

	for _, v := range serviceBeanArray {
		ProxyClient(v, getClientFunc)
	}
}

//定义一个服务发现服务端
func EnableDiscoveryService(consulAddress string, serviceBeans map[string]interface{}, server_address string, server_port int, duration time.Duration, deferFunc func(recover interface{}) string) {

	//注册Rpc服务
	var funcs = []func(){}
	for _, v := range serviceBeans {
		serviceName := reflect.TypeOf(v).Elem().Name()
		//轮询注册 服务发现
		var serviceId = serviceName + ":" + strconv.Itoa(server_port)
		var reg = CreateAgentServiceRegistration(TCP, serviceId, serviceName, server_address, server_port, fmt.Sprint(duration.Seconds()))
		var client = CreateConsulApiClient(consulAddress)
		funcs = append(funcs, func() {
			DoRegister(reg, client)
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

func deleteClient(serviceName string, rpcClient *RpcClient) {
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
