package main

import (
	"net/http"

	"go-to-do-app/server/router"
	"go-to-do-app/util/logger"
)

var log = logger.GetLogger()

func main() {

	r := router.Router()
	log.Info().Msg("Starting server on the port 8080...")

	err := http.ListenAndServe(":8080", r)
	log.Fatal().Err(err).Msg("")
}
