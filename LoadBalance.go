package easyrpc_discovery

import (
	"fmt"
	"hash/crc32"
	"math/rand"
)

type RpcLoadBalanceClient struct {
	Index         int
	rpcClients    []*RpcClient
	RpcClientsMap map[string]*RpcClient
}

func (this *RpcLoadBalanceClient) SetRpcClients(arg []*RpcClient) {
	this.rpcClients = arg
}

func (this RpcLoadBalanceClient) New() RpcLoadBalanceClient {
	this.rpcClients = make([]*RpcClient, 0)
	this.RpcClientsMap = make(map[string]*RpcClient)
	return this
}

func (this *RpcLoadBalanceClient) Append(client RpcClient) {
	if this.rpcClients == nil || this.RpcClientsMap == nil {
		*this = RpcLoadBalanceClient{}.New()
	}
	this.RpcClientsMap[client.Address] = &client
	this.rpcClients = make([]*RpcClient, 0)
	for _, v := range this.RpcClientsMap {
		this.rpcClients = append(this.rpcClients, v)
	}
}

func (this *RpcLoadBalanceClient) Delete(address string, DeleteFunc func(client *RpcClient)) {
	if this.rpcClients == nil || this.RpcClientsMap == nil {
		*this = RpcLoadBalanceClient{}.New()
	}
	if DeleteFunc != nil {
		DeleteFunc(this.RpcClientsMap[address])
	}
	delete(this.RpcClientsMap, address)
	this.rpcClients = make([]*RpcClient, 0)
	for _, v := range this.RpcClientsMap {
		this.rpcClients = append(this.rpcClients, v)
	}
}

func (this *RpcLoadBalanceClient) DeleteAll(DeleteFunc func(client *RpcClient)) {
	if this.rpcClients == nil || this.RpcClientsMap == nil {
		*this = RpcLoadBalanceClient{}.New()
	}
	for _, v := range this.RpcClientsMap {
		if DeleteFunc != nil {
			DeleteFunc(v)
		}
	}
	this.RpcClientsMap = make(map[string]*RpcClient)
	this.rpcClients = make([]*RpcClient, 0)
}

type LoadBalanceType int

const (
	LoadBalanceType_Round  LoadBalanceType = iota //加权轮询
	LoadBalanceType_Random                        //随机
	LoadBalanceType_HASH                          //源地址哈希法
)

//负载均衡实现类
//目前实现 随机，加权轮询,源地址哈希法
func DoBalance(requestIp string, balanceClient *RpcLoadBalanceClient, balanceType *LoadBalanceType) *RpcClient {
	if balanceClient == nil || len(balanceClient.rpcClients) == 0 {
		return nil
	}
	if balanceType == nil {
		//default
		var def = LoadBalanceType_Round
		balanceType = &def
	}
	if *balanceType == LoadBalanceType_Random {
		return randomPickClient(balanceClient)
	} else if *balanceType == LoadBalanceType_Round {
		return roundPickClient(balanceClient)
	} else if *balanceType == LoadBalanceType_HASH {
		return hashPickClient(balanceClient, requestIp)
	} else {
		return nil
	}
}

//源地址哈希 进行负载均衡，相同的IP客户端，如果服务器列表不变，将映射到同一个后台服务器进行访问。
func hashPickClient(client *RpcLoadBalanceClient, requestClientIp string) *RpcClient {
	var defKey string
	if len(requestClientIp) > 0 {
		defKey = requestClientIp
	} else {
		defKey = fmt.Sprintf("%d", rand.Int())
	}
	length := len(client.rpcClients)
	if length == 0 {
		fmt.Errorf("No backend instance")
		return nil
	}
	crcTable := crc32.MakeTable(crc32.IEEE)
	hashValue := crc32.Checksum([]byte(defKey), crcTable)
	index := int(hashValue) % length
	return client.rpcClients[index]
}

//轮询 将请求按顺序轮流分配到后台服务器上，均衡的对待每一台服务器，而不关心服务器实际的连接数和当前的系统负载。
func roundPickClient(client *RpcLoadBalanceClient) *RpcClient {
	var returnObj = client.rpcClients[client.Index]
	var length = len(client.rpcClients)
	client.Index = client.Index + 1
	if client.Index >= length {
		client.Index = 0
	}
	return returnObj
}

//随机选取 通过系统随机函数，根据后台服务器列表的大小值来随机选取其中一台进行访问
func randomPickClient(client *RpcLoadBalanceClient) *RpcClient {
	var length = len(client.rpcClients)
	var randIndex = rand.Intn(length)
	if randIndex < length {
		return client.rpcClients[randIndex]
	}
	return nil
}
