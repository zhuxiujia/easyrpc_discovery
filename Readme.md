
# 基于easyrpc 定制的对接consul服务发现
* 自带负载均衡算法 随机 加权轮询 源地址哈希法
* 基于easyrpc,类似标准库的api，定义服务没有标准库的要求那么严格（可选不传参数，或者只有一个参数，只有一个返回值） https://github.com/zhuxiujia/easyrpc
* 基于easyrpc，负载均衡算法，失败重试，支持动态代理，支持GoMybatis事务，AOP代理，事务嵌套，tag定义事务
![Image text](https://zhuxiujia.github.io/gomybatis.io/assets/easy_consul.png)


* 使用方法
```
go get github.com/zhuxiujia/easyrpc_discovery
go get https://github.com/zhuxiujia/easy_discovery_consul
```
* 下载consul（为什么使用consul，支持健康检查，kv，注册中心）最新客户端 https://www.consul.io/
```
* 启动consul服务注册中心
```
//下载consul最新版，linux版(可选)
./consul agent -dev  -client 0.0.0.0 -ui
```
或者
```
//下载consul最新版，windows版(可选)
consul.exe agent -dev  -client 0.0.0.0 -ui
```
如果以上不满足，可以自定义实现Register接口和（注册到注册中心）ServiceFetcher接口（获取服务列表）来支持更多的注册中心


* 首先定义服务
``` go
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
```

* 首先注册服务端
``` go
var act = TestService{}.New()

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
	easyrpc_discovery.EnableDiscoveryService(services, address, port, 5*time.Second, deferFunc, func() easyrpc_discovery.Register {
		return &easy_discovery_consul.ConsulManager{ConsulAddress: consul}
	})

```

* 首先注册客户端
``` go
    var consulManager = easy_discovery_consul.ConsulManager{ConsulAddress: "127.0.0.1:8500"}
	var act TestService
	easyrpc_discovery.EnableDiscoveryClient(nil, "TestApp", "127.0.0.1", 8500, 5*time.Second, &easyrpc_discovery.RpcConfig{
		RetryTime: 1,
	}, []easyrpc_discovery.RpcServiceBean{
		{
			Service:           &act,
			ServiceName:       "TestService",
			RemoteServiceName: "TestService",
		},
	}, &consulManager, &consulManager)

```
* 测试远程调用
``` go
    for i := 0; i < 5; i++ {
		client.AddActivity(TestVO{
			Name: "test",
		})
		time.Sleep(time.Second)
	}
```
* 如果以上配置正确，打开浏览器 http://localhost:8500 可以看到服务启动成功,然后即可访问微服务了（client.AddActivity（）....更多）


![Image text](https://zhuxiujia.github.io/gomybatis.io/assets/consul_admin.png)

* TODO 
未来会支持更多注册中心，etcd
