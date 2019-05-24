package test

import (
	"encoding/json"
	"fmt"
	"github.com/zhuxiujia/easyrpc_discovery"
	"testing"
	"time"
)

type TestVO struct {
	Name       string     `json:"name"`
	CreateTime *time.Time `json:"create_time"`
}

type TestService struct {
	AddActivity func(arg TestVO) (TestVO, error)
}

func (it TestService) New() TestService {
	it.AddActivity = func(arg TestVO) (result TestVO, e error) {
		var d, _ = json.Marshal(arg)
		println("arg:", string(d))                       //打印远程参数
		println("arg:", string(arg.CreateTime.String())) //打印远程参数
		result.Name = "ffff"
		return result, nil
	}
	return it
}

func TestEnableDiscoveryService(t *testing.T) {
	go registerServer()
	time.Sleep(time.Second)
	var client = registerClient()

	now, _ := time.Parse("2006-01-02 15:04:05", "2019-05-22 15:13:55")
	for i := 0; i < 5; i++ {
		var r, e = client.AddActivity(TestVO{
			Name:       "test",
			CreateTime: &now,
		})
		time.Sleep(time.Second)
		if e != nil {
			println(e.Error())
		}
		println("done:", r.Name)
	}
}

func registerClient() *TestService {
	var act TestService
	easyrpc_discovery.EnableDiscoveryClient(nil, "127.0.0.1:8500", "TestApp", "127.0.0.1", 8500, 5*time.Second, &easyrpc_discovery.RpcConfig{
		RetryTime: 1,
	}, []easyrpc_discovery.RpcServiceBean{
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
	var address = "127.0.0.1"
	var consul = "127.0.0.1:8500"
	var port = 8098

	var services = make(map[string]interface{}, 0)
	services["TestService"] = &act
	var deferFunc = func(recover interface{}) string {
		return fmt.Sprint(recover)
	}
	easyrpc_discovery.EnableDiscoveryService(consul, services, address, port, 5*time.Second, deferFunc)
}
