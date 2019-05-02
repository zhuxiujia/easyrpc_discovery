package easyrpc_discovery

import (
	"errors"
	"github.com/zhuxiujia/easyrpc"
	"github.com/zhuxiujia/easyrpc/easy_jsonrpc"
	"log"
)

type RpcConnectionFactory struct {
}

//新开一个连接
func (factory RpcConnectionFactory) GetConnection(serviceName string, url string) (*easyrpc.Client, error) {
	if url == `` {
		log.Println("[RpcConnectionFactory] connecting rpc fail:" + serviceName + "," + url)
		return nil, errors.New("[RpcConnectionFactory] connecting rpc fail:" + serviceName + "," + url)
	}
	log.Println("[RpcConnectionFactory] connecting rpc:" + serviceName + "," + url + " ...")
	rpcClient, e := easy_jsonrpc.Dial("tcp", url)
	if e != nil || rpcClient == nil {
		log.Println("[RpcConnectionFactory] connecting rpc fail:"+serviceName+","+url, e)
		return nil, e
	} else {
		log.Println("[RpcConnectionFactory] connecting rpc:" + serviceName + "," + url + " success")
	}
	return rpcClient, nil
}

//关闭连接
func (factory RpcConnectionFactory) Close(conn *easyrpc.Client) {
	if conn != nil {
		(*conn).Close()
	}
}
