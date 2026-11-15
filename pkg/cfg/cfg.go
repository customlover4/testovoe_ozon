package cfg

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Web struct {
	Addr              string        `yaml:"addr" env:"ADDR" env-default:"127.0.0.1"`
	Port              string        `yaml:"port" env:"PORT" env-default:"8080"`
	ReadTimeout       time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"READ_HEADER_TIMEOUT" env-default:"3s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
}

type Handler struct {
	Debug bool `yaml:"debug"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Auth struct {
	Secret string `yaml:"secret" env:"SECRET" env-required:""`
}

type Config struct {
	ConnString string `yaml:"db_conn"`
	Handler    `yaml:"handler"`
	Web        `yaml:"web"`
	Logger     `yaml:"logger"`
	Auth       `yaml:"auth"`
}

func MustLoad(path string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
