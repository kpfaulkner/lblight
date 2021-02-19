package main

import (
	"fmt"
	"github.com/kpfaulkner/lblight/pkg"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func initLogging(logFile string) {
	var file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could Not Open Log File : " + err.Error())
	}
	log.SetOutput(file)
	log.SetFormatter(&log.TextFormatter{})
}

func registerPaths(lbl *pkg.LBLight, config pkg.Config) {

	for _, beConfig := range config.BackendRouterConfigs {
		pathMap := make(map[string]bool)
		for _, path := range beConfig.AcceptedPaths {
			pathMap[path] = true
		}

		// make header map later.
		//headerMap := make(map[string]map[string])
		ber := pkg.NewBackendRouter(nil, pathMap)

		// now add backends that the router will route to.
		for _, bec := range beConfig.BackendConfigs {
			be := pkg.NewBackend(bec.Host, bec.Port, bec.MaxConnections)
			ber.AddBackend(be)
		}

		lbl.AddBackendRouter(ber)
	}
}

func main() {

	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	initLogging("lblight.log")
	var config pkg.Config

	var port int
	portStr := os.Getenv("HTTP_PLATFORM_PORT")
	if portStr == "" {
		config = pkg.LoadConfig("lblight.json")
		port = config.Port
	} else {
		port, _ = strconv.Atoi(portStr)
		config = pkg.LoadConfig("d:/home/site/wwwroot/lblight.json")
	}

	log.Infof("port is %d", port)
	lbl := pkg.NewLBLight(port, config.TlsListener)

	registerPaths(lbl, config)

	err := lbl.ListenAndServeTraffic(config.CertCrtPath, config.CertKeyPath)
	if err != nil {
		log.Fatalf("LBLight exiting with error %s", err.Error())
	}
}
