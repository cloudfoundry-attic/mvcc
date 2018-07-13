package config

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	file *os.File
}

func Write(opts ...Option) (*Config, error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	file, err := ioutil.TempFile("", "config")
	if err != nil {
		return nil, err
	}

	y, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	if _, err = file.Write(y); err != nil {
		return nil, err
	}

	return &Config{
		file: file,
	}, nil
}

func (c *Config) Remove() error {
	return os.Remove(c.file.Name())
}

func (c *Config) Name() string {
	return c.file.Name()
}
