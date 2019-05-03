package test

import (
	"encoding/json"
	"github.com/zhuxiujia/easyrpc_discovery"
	"testing"
	"time"
)

type TestVO struct {
	Name string `json:"name"`
}

type TestService struct {
	AddActivity func(arg TestVO, result *TestVO) error
}

func (it TestService) New() TestService {
	it.AddActivity = func(arg TestVO, result *TestVO) error {
		var d, _ = json.Marshal(arg)
		println("arg:", string(d)) //打印远程参数
		result.Name = "ffff"
		return nil
	}
	return it
}

func TestEnableDiscoveryService(t *testing.T) {

	go registerServer()
	time.Sleep(time.Second)

	var client = registerClient()
	for i := 0; i < 5; i++ {
		var r = TestVO{}

		var e = client.AddActivity(TestVO{
			Name: "test",
		}, &r)
		time.Sleep(time.Second)
		if e != nil {
			println(e.Error())
		}
		println("done:", r.Name)
	}
}

func registerClient() *TestService {
	var act TestService
	easyrpc_discovery.EnableDiscoveryClient("127.0.0.1:8500", "TestApp", "127.0.0.1", 8500, 5*time.Second, &easyrpc_discovery.RpcConfig{
		RetryTime: 1,
	}, []easyrpc_discovery.RpcServiceBean{
		{
			Service:           &act,
			ServiceName:       "TestService",
			RemoteServiceName: "TestCoreService",
		},
	}, true)
	return &act
}

func registerServer() {
	var act = TestService{}.New()

	//远程服务信息
	var service = "TestCoreService"
	var address = "127.0.0.1"
	var consul = "127.0.0.1:8500"
	var port = 8098

	var services = make(map[string]interface{}, 0)
	services["TestService"] = &act
	easyrpc_discovery.EnableDiscoveryService(consul, service, services, address, port, 5*time.Second)
}
