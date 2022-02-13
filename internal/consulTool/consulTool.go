package consultool

import (
	"context"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net/http"
	"strconv"
	"time"

	consul "github.com/hashicorp/consul/api"
)

var (
	C        *consul.Agent
	services map[string]*consul.AgentService
)

func NewConsulClient(ip string, port int) (err error) {
	config := consul.DefaultConfig()
	config.Address = ip + ":" + strconv.Itoa(port)
	c, err := consul.NewClient(config)
	if err != nil {
		return
	}
	C = c.Agent()
	return
}

func Register(
	name string,
	ip string,
	port int,
	tags string,
	checkInterval string,
	checkTimeout string,
	checkUrl string) error {
	reg := &consul.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Address: ip,
		Port:    port,
		Tags:    []string{tags},
		Check: &consul.AgentServiceCheck{
			Interval: checkInterval,
			Timeout:  checkTimeout,
			HTTP:     checkUrl,
		},
	}
	return C.ServiceRegister(reg)
}

func DeRegister(name string) error {
	return C.ServiceDeregister(name)
}

func getServerService(filter string) (map[string]*consul.AgentService, error) {
	return C.ServicesWithFilter(filter)
}

// blocking execution
func HealthCheckHttpServer(healthCheckPort int, healthPath string) error {
	http.HandleFunc(healthPath, func(rw http.ResponseWriter, r *http.Request) {})
	return http.ListenAndServe("0.0.0.0:"+strconv.Itoa(healthCheckPort), nil)
}

func ScheduleHealthCheck(
	ctx context.Context,
	addressChannelMap map[string]chan *innerData.InnerDataTransfer,
	addressChannel chan string,
	filter string) (err error) {

	ticker := time.NewTicker(8 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			services, err = getServerService(filter)
			if err != nil {
				return
			}

			for _, v := range services {
				address := v.Address + ":" + strconv.Itoa(v.Port)
				if _, ok := addressChannelMap[address]; ok {
					continue
				} else {
					addressChannel <- address
				}
			}
		}
	}
}
