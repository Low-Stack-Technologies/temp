package env

import (
	_ "github.com/joho/godotenv/autoload"
	"tech.low-stack.temp/shared/env_utils"
)

var (
	ServiceUrl string
)

func LoadVariables() {
	ServiceUrl = env_utils.GetEnvStringWithDefault("TEMP_SERVICE_URL", "http://localhost:8080")
}
