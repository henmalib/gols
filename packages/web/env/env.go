package env

import (
	"os"
)

type EnvType struct {
	Host   string
	ApiKey string
}

func GetEnv() EnvType {
	env := EnvType{}

	env.Host = os.Getenv("HOST")
	env.ApiKey = os.Getenv("API_KEY")

	if env.Host == "" {
		env.Host = ":5050"
	}
	if env.ApiKey == "" {
		panic("No API_KEY provided")
	}

	return env
}

var Env = GetEnv()
