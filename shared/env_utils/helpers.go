package env_utils

import (
	"os"
	"strconv"
	"time"
)

func GetEnvString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Environment variable " + key + " is not set")
	}

	return value
}

func GetEnvStringWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func GetEnvInt(key string) int {
	strValue := GetEnvString(key)
	value, err := strconv.Atoi(strValue)
	if err != nil {
		panic("Environment variable " + key + " is not a valid integer")
	}

	return value
}

func GetEnvDirPath(key string, mustExist bool) string {
	value := GetEnvString(key)
	file, err := os.Stat(value)

	// If the file does not exist and mustExist is true, panic
	if os.IsNotExist(err) && mustExist {
		panic("Environment variable " + key + " points to a non-existing directory")
	} else if err != nil && !os.IsNotExist(err) {
		panic("Failed to check if environment variable " + key + " points to a directory")
	}

	// If the file does exist, check if it is a directory
	if !os.IsNotExist(err) && !file.IsDir() {
		panic("Environment variable " + key + " does not point to a directory")
	}

	return value
}

func GetEnvFilePath(key string, mustExist bool) string {
	value := GetEnvString(key)
	file, err := os.Stat(value)

	// If the file does not exist and mustExist is true, panic
	if os.IsNotExist(err) && mustExist {
		panic("Environment variable " + key + " points to a non-existing file")
	} else if err != nil && !os.IsNotExist(err) {
		panic("Failed to check if environment variable " + key + " points to a file")
	}

	// If the file does exist, check if it is a file
	if !os.IsNotExist(err) && !file.Mode().IsRegular() {
		panic("Environment variable " + key + " does not point to a file")
	}

	return value
}

func GetEnvDuration(key string) time.Duration {
	strValue := GetEnvString(key)
	duration, err := time.ParseDuration(strValue)
	if err != nil {
		panic("Environment variable " + key + " is not a valid duration")
	}

	return duration
}
