package helper

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Network struct {
	IP              string
	LBPort          int    `yaml:"lbPort"`
	HealthPort      int    `yaml:"healthPort"`
	PrometheusPort  int    `yaml:"prometheusPort"`
	TraceServerPort int    `yaml:"traceServerPort"`
	HealthCheckPath string `yaml:"healthCheckPath"`
}

type Consul struct {
	Check struct {
		Interval           string
		Timeout            string
		HttpHealthCheckUrl string `yaml:"httpHealthCheckUrl"`
	}
	Server struct {
		IP   string
		Port int
	}
	Tags   string
	Filter string
}

type Size struct {
	InfoChannelSize       int `yaml:"infoChannelSize"`
	DataChannelSize       int `yaml:"dataChannelSize"`
	LBChannelSize         int `yaml:"lbChannelSize"`
	ConnectionChannelSize int `yaml:"connectionChannelSize"`
}

type Trace struct {
	BatchSize int `yaml:"batchSize"`
}

type Config struct {
	N Network `yaml:"network"`
	C Consul  `yaml:"consul"`
	S Size    `yaml:"size"`
	T Trace   `yaml:"trace"`
}

func ReadConf(filename string) (c *Config, err error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c = &Config{}
	err = yaml.Unmarshal(buf, c)
	return
}
