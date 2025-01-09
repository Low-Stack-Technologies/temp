package env

import _ "github.com/joho/godotenv/autoload"

var (
	HttpPort     int
	StoragePath  string
	DatabasePath string
	BaseUrl      string
)

func LoadVariables() {
	HttpPort = getEnvInt("HTTP_PORT")
	StoragePath = getEnvDirPath("STORAGE_PATH", true)
	DatabasePath = getEnvFilePath("DATABASE_PATH", false)
	BaseUrl = getEnvString("BASE_URL")
}
