package env

import (
	_ "github.com/joho/godotenv/autoload"
	"tech.low-stack.temp/shared/env_utils"
	"time"
)

var (
	HttpPort     int
	DatabasePath string
	BaseUrl      string

	StoragePath  string
	MinFreeSpace uint64

	DefaultExpiration time.Duration
	MaxExpiration     time.Duration
	MinExpiration     time.Duration
)

func LoadVariables() {
	HttpPort = env_utils.GetEnvInt("HTTP_PORT")
	DatabasePath = env_utils.GetEnvFilePath("DATABASE_PATH", false)
	BaseUrl = env_utils.GetEnvString("BASE_URL")

	StoragePath = env_utils.GetEnvDirPath("STORAGE_PATH", true)
	MinFreeSpace = env_utils.GetEnvSize("MIN_FREE_SPACE")

	DefaultExpiration = env_utils.GetEnvDuration("DEFAULT_EXPIRATION")
	MaxExpiration = env_utils.GetEnvDuration("MAX_EXPIRATION")
	MinExpiration = env_utils.GetEnvDuration("MIN_EXPIRATION")
}
