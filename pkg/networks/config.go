package networks

import (
	"golang.org/x/time/rate"
)

type Config struct {
	Networks []*NetworkConfig `yaml:"networks"`
}

type NetworkConfig struct {
	Name  string        `yaml:"name"`
	Type  string        `yaml:"type"`
	Hosts []*HostConfig `yaml:"hosts"`
}
type HostConfig struct {
	URL        string     `yaml:"url"`
	BurstLimit int        `yaml:"burst_limit,omitempty"`
	RateLimit  rate.Limit `yaml:"rate_limit,omitempty"`
	BatchSize  int        `yaml:"batch_size,omitempty"`
}
