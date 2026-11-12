package cfg

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Web struct {
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
}

type Handler struct {
	Debug bool `yaml:"debug"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Config struct {
	ConnString string `yaml:"db_conn"`
	Handler    `yaml:"handler"`
	Web        `yaml:"web"`
	Logger     `yaml:"logger"`
}

func MustLoad(path string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
