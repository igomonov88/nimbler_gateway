package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Web    Web    `yaml:"web"`
	Zipkin Zipkin `yaml:"zipkin"`
	Server Server `yaml:"server"`
}

type Web struct {
	APIHost         string        `yaml:"api_host"`
	DebugHost       string        `yaml:"debug_host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Zipkin struct {
	LocalEndpoint string  `yaml:"local_endpoint"`
	ReporterURI   string  `yaml:"reporter_uri"`
	ServiceName   string  `yaml:"service_name"`
	Probability   float64 `yaml:"probability"`
}

type Server struct {
	APIHost         string        `yaml:"api_host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func Parse(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file path")
	}
	defer file.Close()
	var cfg Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "decoding config file")
	}
	return &cfg, nil
}
