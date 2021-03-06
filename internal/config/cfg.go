package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Addr      string `mapstructure:"addr"`
	TgBotAddr string `mapstructure:"tg_bot_addr"`

	PermissionAddr string           `mapstructure:"permission_addr"`
	ProxyServices  []ServicesConfig `mapstructure:"proxy_services"`

	Jaeger         JaegerConfig `mapstructure:"jaeger"`
	Consul         ConsulConfig `mapstructure:"consul"`
	SwaggerBaseURL string       `mapstructure:"swagger_base_url"`
	Logger         LoggerConfig `mapstructure:"logger"`
}

type ServicesConfig struct {
	Addr      string            `mapstructure:"addr"`
	Endpoints []EndpointsConfig `mapstructure:"endpoints"`
}

type EndpointsConfig struct {
	URI    string `mapstructure:"uri"`
	Method string `mapstructure:"method"`
}

type JaegerConfig struct {
	AgentAddr   string `mapstructure:"agent_addr"`
	ServiceName string `mapstructure:"service_name"`
}

type ConsulConfig struct {
	Addr              string `mapstructure:"addr"`
	AgentAddr         string `mapstructure:"agent_addr"`
	ServiceFamilyName string `mapstructure:"service_family_name"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// Load create configuration from file & environments.
func Load(path string) (*Config, error) {
	dir, file := filepath.Split(path)
	viper.SetConfigName(strings.TrimSuffix(file, filepath.Ext(file)))
	viper.AddConfigPath(dir)

	var cfg Config

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file, %w", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("fail to decode into struct, %w", err)
	}

	return &cfg, nil
}
