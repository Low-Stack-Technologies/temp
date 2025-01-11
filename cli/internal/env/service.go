package env

import (
	_ "github.com/joho/godotenv/autoload"
	"tech.low-stack.temp/shared/env_utils"
)

var (
	ServiceUrl  string
	ReleasesUrl string
)

func LoadVariables() {
	ServiceUrl = env_utils.GetEnvStringWithDefault("TEMP_SERVICE_URL", "https://temp.low-stack.tech")
	ReleasesUrl = env_utils.GetEnvStringWithDefault("TEMP_RELEASES_URL", "https://api.github.com/repos/low-stack-technologies/temp/releases")
}
