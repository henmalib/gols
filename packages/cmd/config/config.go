package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/henmalib/gols/packages/cmd/utils"
)

type Config struct {
	Server  string `json:"server" validate:"required,uri"`
	AuthKey string `json:"authkey" validate:"required"`
}

func resolveConfigFolder() (string, error) {
	configPath, err := os.UserConfigDir()

	if err != nil {
		return "", fmt.Errorf("Couldn't find config dir: %s", err.Error())
	}

	return path.Join(configPath, "gols"), nil
}

func resolveConfigPath() (string, error) {
	configPath, err := resolveConfigFolder()
	if err != nil {
		return "", nil
	}

	return path.Join(configPath, "config.json"), nil
}

func ReadConfigFile() (Config, error) {
	configPath, err := resolveConfigPath()
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(configPath)

	cfg := Config{}
	if err != nil {
		return cfg, err
	}

	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return cfg, nil
}

func WriteConfigFile(data *Config) error {
	if err := utils.Validate.Struct(data); err != nil {
		return err
	}

	configFolder, err := resolveConfigFolder()
	if err != nil {
		return err
	}

	_ = os.Mkdir(configFolder, 0700)

	configPath, err := resolveConfigPath()
	if err != nil {
		return err
	}

	configString, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, configString, 0700)

}
