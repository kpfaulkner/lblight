package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var severCount = 0

const (
	SERVER1 = "https://23.101.125.207:5001"
	PORT    = "443"
)

func initLogging(logFile string) {
	var file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could Not Open Log File : " + err.Error())
	}
	log.SetOutput(file)
	log.SetFormatter(&log.TextFormatter{})
}

func dialTLS(network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{ServerName: host, InsecureSkipVerify: true}

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	cs := tlsConn.ConnectionState()
	cert := cs.PeerCertificates[0]

	// Verify here
	cert.VerifyHostname(host)
	log.Infof(fmt.Sprintf("%v", cert.Subject))

	return tlsConn, nil
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {

	fmt.Printf("serveReverseProxy start\n")
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = &http.Transport{DialTLS: dialTLS}

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)

	fmt.Printf("serveReverseProxy end\n")
}

// Log the typeform payload and redirect url
func logRequestPayload(proxyURL string) {
	log.Infof("proxy_url: %s\n", proxyURL)
}

// Balance returns one of the servers based using round-robin algorithm
func getProxyURL() string {
	var servers = []string{SERVER1}

	server := servers[severCount]
	severCount++

	// reset the counter and start from the beginning
	if severCount >= len(servers) {
		severCount = 0
	}

	return server
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	log.Infof("have req")

	url := getProxyURL()

	log.Infof("Proxy url is %s", url)

	logRequestPayload(url)

	serveReverseProxy(url, res, req)
}

func main() {
	// start server
	/*
		http.HandleFunc("/all/users", handleRequestAndRedirect)
		log.Fatal(http.ListenAndServe(":"+PORT, nil)) */

	initLogging("d:/home/LogFiles/loadbalancer.log")
	port := os.Getenv("HTTP_PLATFORM_PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("port is %s", port)

	err := http.ListenAndServeTLS(":"+port, "d:/home/site/wwwroot/localhost.crt", "d:/home/site/wwwroot/localhost.key", http.HandlerFunc(handleRequestAndRedirect))
	if err != nil {
		log.Errorf("error %s\n", err.Error())
	}

}
