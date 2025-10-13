package config

import "sync"

type HostConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Scheme   string `yaml:"scheme"`
	Hostname string
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type RootConfig struct {
	Mutex         sync.Mutex
	Address       string                 `yaml:"address"`
	Port          uint                   `yaml:"port"`
	HttpsProxy    string                 `yaml:"https_proxy"`
	MetricsPrefix string                 `yaml:"metrics_prefix"`
	TLS           TLSConfig              `yaml:"tls"`
	Timeout       uint                   `yaml:"timeout"`
	Hosts         map[string]*HostConfig `yaml:"hosts"`
}
