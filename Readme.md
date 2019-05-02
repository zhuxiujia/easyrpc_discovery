
# 基于easyrpc 定制的对接consul服务发现
* 自带负载均衡算法 随机 加权轮询 源地址哈希法
* 基于easyrpc,类似标准库的api，定义服务没有标准库的要求那么严格（可选不传参数，或者只有一个参数，只有一个返回值）
* 基于easyrpc，负载均衡算法，失败重试，支持动态代理，支持GoMybatis事务，AOP代理，事务嵌套，tag定义事务
![Image text](https://zhuxiujia.github.io/gomybatis.io/assets/easy_consul.png)


* 使用方法
```
go get github.com/zhuxiujia/easyrpc_discovery
go get github.com/zhuxiujia/easyrpc
go get github.com/hashicorp/consul
```

* 首先定义服务
```
type ActivityService struct {
	AddActivity func(arg vo.ActivityVO) error
}

func (it ActivityService) New() ActivityService {
	it.AddActivity = func(arg vo.ActivityVO) error {
		var d, _ = json.Marshal(arg)
		println("add activity", string(d)) //打印远程参数
		return errors.New("fuck error")
	}
	return it
}
```

* 首先注册服务端
```
var act = service.ActivityService{}.New()

	//远程服务信息
	var service = "ActivityService"
	var address = "127.0.0.1"
	var consul = "127.0.0.1:8500"
	var port = 1234

	var services = make(map[string]interface{}, 0)
	services["ActivityService"] = &act
	easyrpc_discovery.EnableDiscoveryService(consul, service, services, address, port, 5*time.Second)

```

* 首先注册客户端
```
easyrpc_discovery.EnableDiscoveryClient("127.0.0.1:8500", "App", "127.0.0.1", 8500, 5*time.Second, &easyrpc_discovery.RpcConfig{
		RetryTime: 1,
	}, []easyrpc_discovery.RpcServiceBean{
		{
			Service:           &act,
			RemoteServiceName: "ActivityService",
		},
	}, true)

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
* TODO 
未来会支持更多注册中心，etcd
