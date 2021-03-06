package easyrpc_discovery

import (
	"reflect"
	"strings"
)

//UseService 可写入每个函数代理方法
func ProxyClient(bean RpcServiceBean, GetClient func(RemoteServiceName string) (*RpcClient, error)) {
	v := reflect.ValueOf(bean.Service)
	if v.Kind() != reflect.Ptr {
		panic("UseService: remoteService argument must be a pointer")
	}
	Proxy(bean.Service, func(funcField reflect.StructField, field reflect.Value) func(arg ProxyArg) []reflect.Value {

		var returnType = makeReturnType(funcField)

		return func(arg ProxyArg) []reflect.Value {
			var rpcClient, e = GetClient(bean.RemoteServiceName)
			//build ptr
			var returnV reflect.Value
			var result interface{}
			if returnType.ReturnOutType != nil {
				returnV = reflect.New(returnType.ReturnOutType)
				switch (returnType.ReturnOutType).Kind() {
				case reflect.Map:
					returnV.Elem().Set(reflect.MakeMap(returnType.ReturnOutType))
				case reflect.Slice:
					returnV.Elem().Set(reflect.MakeSlice(returnType.ReturnOutType, 0, 0))
				}
				result = returnV.Interface()
			}
			var remoteServiceFunc = bean.ServiceName + "." + funcField.Name
			if e != nil {
				return buildReturnValues(&returnType, &returnV, e)
			}
			var callArg interface{}
			if arg.ArgsLen > 0 {
				callArg = arg.Args[0].Interface()
			}
			e = rpcClient.Call(bean.ServiceName, remoteServiceFunc, callArg, result)
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
