package easyrpc_discovery

import (
	"reflect"
)

// UseService 可写入每个函数代理方法
func UseService(remoteService interface{}, proxyFunc func(method string, args []reflect.Value) error, namespace ...string) {
	ns := ""
	if len(namespace) == 1 {
		ns = namespace[0]
	}
	v := reflect.ValueOf(remoteService)
	if v.Kind() != reflect.Ptr {
		panic("UseService: remoteService argument must be a pointer")
	}
	buildRemoteService(v, ns, proxyFunc)
}
func buildRemoteService(v reflect.Value, ns string, proxyFunc func(method string, args []reflect.Value) error) {
	v = v.Elem()
	t := v.Type()
	et := t
	if et.Kind() == reflect.Ptr {
		et = et.Elem()
	}
	ptr := reflect.New(et)
	obj := ptr.Elem()
	count := obj.NumField()
	for i := 0; i < count; i++ {
		f := obj.Field(i)
		ft := f.Type()
		sf := et.Field(i)
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if f.CanSet() {
			switch ft.Kind() {
			case reflect.Struct:
			case reflect.Func:
				buildRemoteMethod(f, ft, sf, ns, proxyFunc)
			}
		}
	}
	if t.Kind() == reflect.Ptr {
		v.Set(ptr)
	} else {
		v.Set(obj)
	}
}

func buildRemoteMethod(f reflect.Value, ft reflect.Type, sf reflect.StructField, ns string, proxyFunc func(method string, args []reflect.Value) error) {
	var fn func(in []reflect.Value) (out []reflect.Value)
	fn = func(args []reflect.Value) (results []reflect.Value) {
		err := proxyFunc(sf.Name, args)
		results = append(results, reflect.ValueOf(&err).Elem())
		return
	}
	if f.Kind() == reflect.Ptr {
		fp := reflect.New(ft)
		fp.Elem().Set(reflect.MakeFunc(ft, fn))
		f.Set(fp)
	} else {
		f.Set(reflect.MakeFunc(ft, fn))
	}
}
