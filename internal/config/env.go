package config

import (
	"os"
	"strconv"
	"strings"
)

func getEnvString(env string, val *string) {
	value := os.Getenv(env)
	if len(value) == 0 {
		return
	}

	*val = value
}

func getEnvBool(env string, val *bool) {
	value := os.Getenv(env)
	if len(value) == 0 {
		return
	}

	switch strings.ToLower(value) {
	case "0", "false":
		*val = false
	default:
		*val = true
	}
}

func getEnvUint(env string, val *uint) {
	s := os.Getenv(env)
	if len(s) == 0 {
		return
	}

	value, err := strconv.ParseUint(s, 10, 0)
	if err == nil {
		*val = uint(value)
	}
}

func (c *RootConfig) FromEnvironment() {
	var username string
	var password string
	var scheme string

	getEnvString("CONFIG_ADDRESS", &c.Address)
	getEnvString("CONFIG_METRICS_PREFIX", &c.MetricsPrefix)
	getEnvString("CONFIG_DEFAULT_USERNAME", &username)
	getEnvString("CONFIG_DEFAULT_PASSWORD", &password)
	getEnvString("CONFIG_DEFAULT_SCHEME", &scheme)
	getEnvString("CONFIG_TLS_CERT_FILE", &c.TLS.CertFile)
	getEnvString("CONFIG_TLS_KEY_FILE", &c.TLS.KeyFile)

	getEnvUint("CONFIG_PORT", &c.Port)
	getEnvUint("CONFIG_TIMEOUT", &c.Timeout)

	getEnvBool("CONFIG_TLS_ENABLED", &c.TLS.Enabled)

	def, ok := c.Hosts["default"]
	if !ok {
		def = &HostConfig{}
	}

	if len(username) > 0 {
		def.Username = username
		ok = true
	}

	if len(password) > 0 {
		def.Password = password
		ok = true
	}

	if len(scheme) > 0 {
		def.Scheme = scheme
		ok = true
	}

	if ok {
		c.Hosts["default"] = def
	}
}
