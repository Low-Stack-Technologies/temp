package main

import (
	"log"

	"tech.low-stack.temp/server/internal/api"
	"tech.low-stack.temp/server/internal/db"
	"tech.low-stack.temp/server/internal/download"
	"tech.low-stack.temp/server/internal/env"
	"tech.low-stack.temp/server/internal/expiration"
	"tech.low-stack.temp/server/internal/upload"
	"tech.low-stack.temp/server/internal/web"
)

func main() {
	log.Println("Welcome to the Temp server!")

	env.LoadVariables()
	db.Initialize()
	expiration.Initialize()

	upload.Initialize()
	download.Initialize()
	web.Initialize()
	api.Initialize()
}
