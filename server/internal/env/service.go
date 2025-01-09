package env

import (
	_ "github.com/joho/godotenv/autoload"
	"tech.low-stack.temp/shared/env_utils"
)

var (
	HttpPort     int
	StoragePath  string
	DatabasePath string
	BaseUrl      string
)

func LoadVariables() {
	HttpPort = env_utils.GetEnvInt("HTTP_PORT")
	StoragePath = env_utils.GetEnvDirPath("STORAGE_PATH", true)
	DatabasePath = env_utils.GetEnvFilePath("DATABASE_PATH", false)
	BaseUrl = env_utils.GetEnvString("BASE_URL")
}
