package easyrpc_discovery

import (
	consulapi "github.com/hashicorp/consul/api"
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

func FetchServiceMap(clientName string, manager *RpcServiceManager, client *consulapi.Client, clearAllClient func(m map[string]*RpcLoadBalanceClient), pool *ConnPool) {
	newServiceList, error := client.Agent().Services()
	if error != nil {
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
					deleteClient(serviceName, client)
				})
			}
			delete(manager.ServiceAddressMap, serviceName)
		}
	}

	for _, v := range newServiceList {
		var addr = v.Address + ":" + strconv.Itoa(v.Port)
		var serviceName = v.Service
		if serviceName == clientName {
			continue
		}
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

func AddOne(manager *RpcServiceManager, remoteService string, address string, createClient func(serviceName string, address string) interface{}, pool *ConnPool, load *RpcLoadBalanceClient) interface{} {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	var rpcLoadBalanceClient = manager.ServiceAddressMap[remoteService]
	var createRpcObject = createClient(remoteService, address)
	rpcLoadBalanceClient.Append(RpcClient{}.New(address, pool, load, manager.RpcConfig.RetryTime))
	manager.ServiceAddressMap[remoteService] = rpcLoadBalanceClient
	return createRpcObject
}
