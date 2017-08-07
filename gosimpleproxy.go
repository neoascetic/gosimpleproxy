package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"strings"
	"net"
	"net/http"
	"net/http/httputil"
)

type portMap map[string]string

func buildServer(pMap portMap, defPort string) *http.Server {
	director := func(req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		port, found := pMap[strings.ToLower(host)]
		if !found {
			port = defPort
		}
		req.URL.Scheme = "http"
		req.URL.Host = "localhost:" + port
		log.Printf("%s -> %s", req.Host, port)
	}
	return &http.Server{
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  3 * time.Minute,
		Handler:      &httputil.ReverseProxy{Director: director},
	}
}

func buildMapAndDefPort(list []string) (portMap, string, error) {
	var defPort = ""
	var pMap = make(portMap)
	for _, mapping := range list {
		host, port, err := net.SplitHostPort(mapping)
		if err != nil {
			return pMap, defPort, err
		}
		pMap[host] = port
		if defPort == "" {
			defPort = port
		}
	}
	return pMap, defPort, nil
}

func main() {
	pMap, defPort, err := buildMapAndDefPort(os.Args[1:])
	if err != nil || len(pMap) == 0 {
		fmt.Printf("Usage: %s domain:port [domain:port ...]", os.Args[0])
		os.Exit(1)
	}
	log.Fatal(buildServer(pMap, defPort).ListenAndServe())
}
