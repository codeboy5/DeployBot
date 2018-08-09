package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {

	log.Info("Beginning Initialization")
	initialize()
	log.Info("Initialization completed successfully")

	http.HandleFunc("/slack/commands/deploy/", deployCommandHandler)
	http.HandleFunc("/slack/interactive/request/", requestHandler)
	http.HandleFunc("/slack/interactive/data-options/", dataOptionsHandler)

	http.HandleFunc("/github/repo/", repoHandler)

	http.HandleFunc("/logs/", logHandler)

	log.Infof("Starting HTTP Server on :%s", Port)

	log.Fatal(http.ListenAndServe(":"+Port, nil))
}