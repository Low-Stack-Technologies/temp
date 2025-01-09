package api

import (
  "fmt"
  "log"
  "net/http"
  "tech.low-stack.temp/server/internal/env"
)

func Initialize() {
  log.Printf("HTTP server started on http://0.0.0.0:%d\n", env.HttpPort)
  panic(http.ListenAndServe(fmt.Sprintf(":%d", env.HttpPort), nil))
}
