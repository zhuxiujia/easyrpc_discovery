package easyrpc_discovery

import (
	"strconv"
	"strings"
	"sync"
)

const ConnectError = "connection is shut down"

type RpcServiceManager struct {
	Mutex             sync.Mutex
	ServiceAddressMap map[string]*RpcLoadBalanceClient //  map[service]map[addr]interface
	RpcConfig         RpcConfig
}

func (RpcServiceManager) New() RpcServiceManager {
	var ServiceManager = RpcServiceManager{
		Mutex:             sync.Mutex{},
		ServiceAddressMap: map[string]*RpcLoadBalanceClient{},
	}
	return ServiceManager
}

type AgentService struct {
	Service string
	Port    int
	Address string
}

func (this *RpcServiceManager) SetNewServiceMap(manager *RpcServiceManager, newServiceList map[string]*AgentService, clearAllClient func(m map[string]*RpcLoadBalanceClient), pool *ConnPool) {
	if newServiceList == nil {
		return
	}
	for k, _ := range newServiceList {
		if !strings.Contains(k, "Service") {
			delete(newServiceList, k)
		}
	}
	if len(newServiceList) == 0 && len(manager.ServiceAddressMap) != 0 {
		clearAllClient(manager.ServiceAddressMap)
		manager.ServiceAddressMap = make(map[string]*RpcLoadBalanceClient)
		return
	}
	for key, oldApi := range newServiceList {
		var serviceName = oldApi.Service
		var newApi = newServiceList[key]
		if newApi == nil && oldApi != nil {
			//clear service
			var rpcLoadBalanceClient = manager.ServiceAddressMap[serviceName]
			if rpcLoadBalanceClient != nil {
				rpcLoadBalanceClient.DeleteAll(func(client *RpcClient) {
					deleteClient(client)
				})
			}
			delete(manager.ServiceAddressMap, serviceName)
		}
	}
	for _, v := range newServiceList {
		var addr = v.Address + ":" + strconv.Itoa(v.Port)
		var serviceName = v.Service
		var rpcLoadBalanceClient = manager.ServiceAddressMap[serviceName]
		if rpcLoadBalanceClient == nil {
			var client = RpcLoadBalanceClient{}.New()
			rpcLoadBalanceClient = &client
		}
		if rpcLoadBalanceClient.RpcClientsMap[addr] == nil {
			rpcLoadBalanceClient.Append(RpcClient{}.New(addr, pool, rpcLoadBalanceClient, manager.RpcConfig.RetryTime))
		}
		manager.ServiceAddressMap[serviceName] = rpcLoadBalanceClient
	}
}

func (this *RpcServiceManager) AddOne(manager *RpcServiceManager, remoteService string, address string, createClient func(serviceName string, address string) interface{}, pool *ConnPool, load *RpcLoadBalanceClient) interface{} {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	var rpcLoadBalanceClient = manager.ServiceAddressMap[remoteService]
	var createRpcObject = createClient(remoteService, address)
	rpcLoadBalanceClient.Append(RpcClient{}.New(address, pool, load, manager.RpcConfig.RetryTime))
	manager.ServiceAddressMap[remoteService] = rpcLoadBalanceClient
	return createRpcObject
}
