package main

import (
	"fmt"
	"github.com/kpfaulkner/lblight/pkg"
	log "github.com/sirupsen/logrus"
	"os"
)


func initLogging(logFile string) {
	var file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could Not Open Log File : " + err.Error())
	}
	log.SetOutput(file)
	log.SetFormatter(&log.TextFormatter{})
}


func main() {

	initLogging("lblight.log")
	port := os.Getenv("HTTP_PLATFORM_PORT")
	if port == "" {
		port = "8080"
	}

	lbl := pkg.NewLBLight(8080)

	pathMap := make(map[string]bool)
	pathMap["/foo"] = true
	ber := pkg.NewBackendRouter("127.0.0.1",8081, nil, pathMap,10)
	lbl.AddBackendRouter(ber)

	err := lbl.ListenAndServeTraffic()
	if err != nil {
		log.Fatalf("LBLight exiting with error %s", err.Error())
	}
}
