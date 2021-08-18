package easyrpc_discovery

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"strconv"
	"strings"
	"time"
)

type ConsulCheckType int

const (
	TCP ConsulCheckType = iota
	HTTP
)

type ConsulManager struct {
	ConsulAddress string
	reg           *consulapi.AgentServiceRegistration
	client        *consulapi.Client
	manager       *RpcServiceManager
	clearFunc     *func(m map[string]*RpcLoadBalanceClient)
	pool          *ConnPool
}

func (s *ConsulManager) InitRegister(
	serviceName string,
	address string,
	port int,
	duration time.Duration) {
	var serviceId = serviceName + ":" + strconv.Itoa(port)
	var client = s.CreateConsulApiClient(s.ConsulAddress)
	var reg = CreateAgentServiceRegistration(TCP, serviceId, serviceName, address, port, fmt.Sprint(duration.Seconds()))
	s.client = client
	s.reg = reg
}

func (s *ConsulManager) DoRegister() {
	s.DoRegisterConsul(s.reg, s.client)
}

func (s *ConsulManager) InitServiceFetcher(
	manager *RpcServiceManager,
	clearFunc func(m map[string]*RpcLoadBalanceClient),
	pool *ConnPool) {
	s.manager = manager
	s.clearFunc = &clearFunc
	s.pool = pool
}

func (s *ConsulManager) DoFetch() {
	var newService map[string]*AgentService
	newServiceList, e := s.client.Agent().Services()
	if e != nil {
		return
	}
	if newServiceList == nil {
		return
	}
	newService = map[string]*AgentService{}
	for k, _ := range newServiceList {
		if !strings.Contains(k, "Service") {
			delete(newServiceList, k)
		}
	}
	for k, v := range newServiceList {
		newService[k] = &AgentService{
			Service: v.Service,
			Port:    v.Port,
			Address: v.Address,
		}
	}
	s.manager.SetNewServiceMap(s.manager, newService, *s.clearFunc, s.pool)
}

func CreateAgentServiceRegistration(consulCheckType ConsulCheckType, id string, serviceName string, address string, port int, time string) *consulapi.AgentServiceRegistration {
	fmt.Println("[ConsulManager]start register consul Rpc Service")
	//创建一个新服务。
	registration := new(consulapi.AgentServiceRegistration)
	registration.Address = address
	registration.Port = port
	registration.ID = id
	registration.Name = serviceName
	registration.Tags = []string{serviceName}

	//增加check。
	check := new(consulapi.AgentServiceCheck)
	if consulCheckType == TCP {
		check.TCP = registration.Address + ":" + strconv.Itoa(registration.Port)
	} else if consulCheckType == HTTP {
		check.HTTP = registration.Address + ":" + strconv.Itoa(registration.Port)
	}
	//设置超时 5s。
	check.Timeout = time + "s"
	//设置间隔 5s。
	check.Interval = time + "s"
	//check失败后30秒删除本服务
	check.DeregisterCriticalServiceAfter = time + "s"
	//注册check服务。
	registration.Check = check
	return registration
}

func (s *ConsulManager) DoRegisterConsul(registration *consulapi.AgentServiceRegistration, client *consulapi.Client) error {
	err := client.Agent().ServiceRegister(registration)
	if err != nil {
		fmt.Println("[ConsulManager]Register Consul Rpc Service error=", err)
	} else {
		fmt.Println("[ConsulManager]Register Consul Rpc Service success.")
	}
	return err
}

func (s *ConsulManager) CreateConsulApiClient(consulAddress string) *consulapi.Client {
	config := consulapi.DefaultConfig()
	config.Address = consulAddress
	client, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Println("[ConsulManager]new consul client error : ", err)
	}
	return client
}
