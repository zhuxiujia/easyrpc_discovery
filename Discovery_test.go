package easyrpc_discovery

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type TestVO struct {
	Name string
}

type TestService struct {
	AddActivity func(arg TestVO) error
}

func (it TestService) New() TestService {
	it.AddActivity = func(arg TestVO) error {
		var d, _ = json.Marshal(arg)
		println("add activity", string(d)) //打印远程参数
		return errors.New("fuck error")
	}
	return it
}

func TestEnableDiscoveryService(t *testing.T) {

	go registerServer()

	var client = registerClient()
	for i := 0; i < 5; i++ {
		client.AddActivity(TestVO{
			Name: "test",
		})
		time.Sleep(time.Second)
	}
}

func registerClient() *TestService {
	var act TestService
	EnableDiscoveryClient("127.0.0.1:8500", "TestApp", "127.0.0.1", 8500, 5*time.Second, &RpcConfig{
		RetryTime: 1,
	}, []RpcServiceBean{
		{
			Service:           &act,
			ServiceName:       "TestService",
			RemoteServiceName: "TestService",
		},
	}, true)
	return &act
}

func registerServer() {
	var act = TestService{}.New()

	//远程服务信息
	var service = "TestService"
	var address = "127.0.0.1"
	var consul = "127.0.0.1:8500"
	var port = 1234

	var services = make(map[string]interface{}, 0)
	services["TestService"] = &act
	EnableDiscoveryService(consul, service, services, address, port, 5*time.Second)
}
