package consultool

import (
	"strconv"

	consul "github.com/hashicorp/consul/api"
)

var (
	C                     *consul.Agent
	interIp               string                    = "172.17.129.202"
	interLBPort           int                       = 11000
	interHealthPort       int                       = 10001
	convinentInterAddress string                    = interIp + ":" + strconv.Itoa(interHealthPort)
	check                 *consul.AgentServiceCheck = &consul.AgentServiceCheck{
		Interval: "10s",
		Timeout:  "1s",
		HTTP:     "http://" + convinentInterAddress + "/health",
	}
)

func NewConsulClient(addr string) (err error) {
	if addr == "nil" {
		addr = interIp + ":" + strconv.Itoa(8500)
	}
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return
	}
	C = c.Agent()
	return
}

func Register(name string, healthPort int) error {
	interHealthPort = healthPort
	reg := &consul.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Address: interIp,
		Port:    interLBPort,
		Tags:    []string{"lb"},
		Check:   check,
	}
	return C.ServiceRegister(reg)
}

func DeRegister(name string) error {
	return C.ServiceDeregister(name)
}

func GetServerService() (map[string]*consul.AgentService, error) {
	return C.ServicesWithFilter("server in Tags")
}
