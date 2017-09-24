package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"
)

type portMap map[string]string
type regexpMap map[*regexp.Regexp]string

func buildProxy(addr string, pMap portMap, rMap regexpMap, defPort string) *http.Server {
	director := func(req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		host = strings.ToLower(host)
		port, found := pMap[host]
		if !found {
			port = defPort
			for pattern, possiblePort := range rMap {
				if pattern.MatchString(host) {
					pMap[host], port = possiblePort, possiblePort
					break
				}
			}
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
		Addr:         addr,
	}
}

func buildMapAndDefPort(mapList []string) (portMap, regexpMap, string, error) {
	var defPort = ""
	var pMap, rMap = make(portMap, len(mapList)), make(regexpMap)
	for _, mapping := range mapList {
		host, port, err := net.SplitHostPort(mapping)
		if err != nil {
			return pMap, rMap, defPort, err
		}
		if strings.Contains(host, "*") {
			pattern := strings.Replace(regexp.QuoteMeta(host), `\*`, ".+", -1)
			rMap[regexp.MustCompile("^" + pattern + "$")] = port
		} else {
			pMap[host] = port
		}
		if defPort == "" {
			defPort = port
		}
	}
	return pMap, rMap, defPort, nil
}

func main() {
	var addr, cert, key string
	flag.StringVar(&addr, "addr", ":http", "which address listen to")
	flag.StringVar(&cert, "cert", "", "path to the certificate file (requires -key)")
	flag.StringVar(&key, "key", "", "path to the key file (requires -cert)")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-addr=[iface]:port] [(-cert=cert -key=key)] domain:port [domain:port ...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println("  domain:port")
		fmt.Println("    \tdomain may contain simple placeholders, such as *.domain.name")
	}
	flag.Parse()
	pMap, rMap, defPort, err := buildMapAndDefPort(flag.Args())
	if err != nil || len(pMap) == 0 || ((cert != "") != (key != "")) {
		flag.Usage()
		os.Exit(1)
	}
	proxy := buildProxy(addr, pMap, rMap, defPort)
	if cert == "" {
		log.Fatal(proxy.ListenAndServe())
	} else {
		log.Fatal(proxy.ListenAndServeTLS(cert, key))
	}
}
