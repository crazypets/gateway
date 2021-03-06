package consul

import (
	"fmt"
	"gateway/internal/config"
	"net"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

type Service struct {
	cfg config.ConsulConfig

	serviceID string

	client *consulapi.Client
}

func NewService(client *consulapi.Client, cfg config.ConsulConfig, serviceID string) *Service {
	return &Service{
		cfg:       cfg,
		serviceID: serviceID,
		client:    client,
	}
}

func (s *Service) Register() error {
	host, port, err := net.SplitHostPort(s.cfg.AgentAddr)
	if err != nil {
		return fmt.Errorf("parse consul agent addr: %w", err)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse consul agent port: %w", err)
	}

	if err = s.client.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      s.serviceID,
		Name:    s.cfg.ServiceFamilyName,
		Port:    p,
		Address: host,
		Check: &consulapi.AgentServiceCheck{
			Interval: "5s",
			Timeout:  "3s",
			HTTP:     fmt.Sprintf("http://%s:%d/health-check", host, p),
		},
	}); err != nil {
		return fmt.Errorf("sign up service via consul: %w", err)
	}

	return nil
}

func (s *Service) Deregister() error {
	if err := s.client.Agent().ServiceDeregister(s.serviceID); err != nil {
		return fmt.Errorf("deregister service in consul: %w", err)
	}

	return nil
}
