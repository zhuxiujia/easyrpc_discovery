package easyrpc_discovery

import (
	"github.com/zhuxiujia/easyrpc"
	"reflect"
	"strings"
)

//UseService 可写入每个函数代理方法
func ProxyClient(bean RpcServiceBean, GetClient func(arg *RpcClient, b RpcServiceBean) error, retry int) {
	v := reflect.ValueOf(bean.Service)
	if v.Kind() != reflect.Ptr {
		panic("UseService: remoteService argument must be a pointer")
	}
	Proxy(bean.Service, func(funcField reflect.StructField, field reflect.Value) func(arg ProxyArg) []reflect.Value {

		var returnType = makeReturnType(funcField)

		return func(arg ProxyArg) []reflect.Value {

			var e error
			var rpcClient RpcClient
			e = GetClient(&rpcClient, bean)

			//build ptr
			var returnV = reflect.New(returnType.ReturnOutType)
			switch (returnType.ReturnOutType).Kind() {
			case reflect.Map:
				returnV.Elem().Set(reflect.MakeMap(returnType.ReturnOutType))
			case reflect.Slice:
				returnV.Elem().Set(reflect.MakeSlice(returnType.ReturnOutType, 0, 0))
			}
			var result = returnV.Interface()

			var remoteServiceName = bean.ServiceName + "." + funcField.Name
			for i := 0; i < (retry + 1); i++ {
				if e != nil {
					return buildReturnValues(&returnType, nil, e)
				}
				if rpcClient.Object == nil {
					continue
				}
				var callArg interface{}
				if arg.ArgsLen > 0 {
					callArg = arg.Args[0].Interface()
				}
				e = rpcClient.Object.(*easyrpc.Client).Call(remoteServiceName, callArg, result)
				if e == nil {
					return buildReturnValues(&returnType, &returnV, e)
				} else if e.Error() == ConnectError {
					println("[easyrpc] " + e.Error())
					rpcClient.Shutdown = true
					var clientErrr = GetClient(&rpcClient, bean)
					if clientErrr != nil {
						e = clientErrr
					}
				} else {
					return buildReturnValues(&returnType, &returnV, e)
				}
			}
			return buildReturnValues(&returnType, &returnV, e)
		}
	})
}

func makeReturnType(funcField reflect.StructField) ReturnType {
	var t = ReturnType{
		ReturnIndex: -1,
	}
	t.NumOut = funcField.Type.NumOut()
	for i := 0; i < funcField.Type.NumOut(); i++ {
		var item = funcField.Type.Out(i)
		if item.Kind() == reflect.Interface && strings.Contains(item.String(), "error") {
			t.ErrorType = item
		} else {
			t.ReturnOutType = item
			t.ReturnIndex = i
		}
	}
	if t.ErrorType == nil {
		panic("[easyrpc]must have return error")
	}
	return t
}

func buildReturnValues(returnType *ReturnType, returnValue *reflect.Value, e error) []reflect.Value {
	var returnValues = make([]reflect.Value, returnType.NumOut)
	for index, _ := range returnValues {
		if index == returnType.ReturnIndex {
			if returnValue != nil {
				returnValues[index] = (*returnValue).Elem()
			}
		} else {
			if e != nil {
				returnValues[index] = reflect.New(returnType.ErrorType)
				returnValues[index].Elem().Set(reflect.ValueOf(e))
				returnValues[index] = returnValues[index].Elem()
			} else {
				returnValues[index] = reflect.Zero(returnType.ErrorType)
			}
		}
	}
	return returnValues
}
