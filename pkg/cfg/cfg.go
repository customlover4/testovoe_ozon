package cfg

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	ConnString string `yaml:"db_conn"`
}

func MustLoad(path string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
