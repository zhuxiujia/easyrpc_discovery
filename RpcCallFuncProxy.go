package easyrpc_discovery

import (
	"github.com/zhuxiujia/easyrpc"
	"reflect"
)

//UseService 可写入每个函数代理方法
func ProxyClient(bean RpcServiceBean, GetClient func(arg *RpcClient) error, retry int) {
	v := reflect.ValueOf(bean.Service)
	if v.Kind() != reflect.Ptr {
		panic("UseService: remoteService argument must be a pointer")
	}
	Proxy(bean.Service, func(funcField reflect.StructField, field reflect.Value) func(arg ProxyArg) []reflect.Value {
		return func(arg ProxyArg) []reflect.Value {
			var result = ""
			var e error
			var rpcClient RpcClient
			e = GetClient(&rpcClient)
			for i := 0; i < (retry + 1); i++ {
				if e != nil {
					return makeErrors(e, funcField)
				}
				if rpcClient.Object == nil {
					continue
				}
				var remoteServiceName = bean.RemoteServiceName + "." + funcField.Name
				e = rpcClient.Object.(*easyrpc.Client).Call(remoteServiceName, arg.Args[0].Interface(), &result)
				if e == nil {
					return makeErrors(e, funcField)
				} else if e.Error() == ConnectError {
					rpcClient.Shutdown = true
					var clientErrr = GetClient(&rpcClient)
					if clientErrr != nil {
						e = clientErrr
					}
				}
			}
			return makeErrors(e, funcField)
		}
	})
}

func makeErrors(e error, funcField reflect.StructField) []reflect.Value {
	var returnValues reflect.Value
	if e != nil {
		returnValues = reflect.ValueOf(e).Convert(funcField.Type.Out(0))
	} else {
		returnValues = reflect.Zero(funcField.Type.Out(0))
	}
	return []reflect.Value{returnValues}
}
