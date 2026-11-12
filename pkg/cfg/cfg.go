package cfg

import "github.com/ilyakaznacheev/cleanenv"

type db struct {
}

type Config struct {
	ConnString    string `yaml:"db_conn"`
	TelegramToken string `yaml:"telegram_token"`
}

func MustLoad(path string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
