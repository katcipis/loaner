package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/katcipis/loaner/api"
	"github.com/katcipis/loaner/loan"
	log "github.com/sirupsen/logrus"
)

// VersionString version of the service
var VersionString = "no version info"

func main() {
	const timeout = 10 * time.Second

	var port int
	var version bool

	flag.BoolVar(&version, "version", false, "show service version and exit")
	flag.IntVar(&port, "port", 8080, "port where the service will be listening to")
	flag.Parse()

	if version {
		fmt.Printf("loaner version: %q\n", VersionString)
		return
	}

	service := api.New(loan.CreatePlan)
	// A global timeout for an http server may not be the best fit
	// for all scenarios. I worked on streaming APIs in the past and
	// the stream can be long lived (both audio/media and also documents like
	// a JSON stream). So a config like that must be used with care to
	// not cause very odd bugs (like streams being cut short automatically).
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      service,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	log.Infof("running loaner service, listening on port %d", port)
	log.Fatal(server.ListenAndServe())
}
