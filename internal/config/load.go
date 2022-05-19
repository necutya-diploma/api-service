package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	devEnv = "dev"
)

func Load(path string) (config Config, err error) {
	err = cleanenv.ReadConfig(fmt.Sprintf("%sconfig.json", path), &config)
	if err != nil {
		return Config{}, err
	}

	if config.ENV == devEnv {
		godotenv.Load()

		err = cleanenv.ReadEnv(&config)
		if err != nil {
			return Config{}, err
		}
	}

	return config, nil
}
