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

type upstreamMap map[string]string
type regexpMap map[*regexp.Regexp]string

func buildProxy(addr string, uMap upstreamMap, rMap regexpMap, defUpstream string) *http.Server {
	director := func(req *http.Request) {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		host = strings.ToLower(host)
		upstream, found := uMap[host]
		if !found {
			upstream = defUpstream
			for pattern, possibleUpstream := range rMap {
				if pattern.MatchString(host) {
					uMap[host], upstream = possibleUpstream, possibleUpstream
					break
				}
			}
		}
		req.URL.Scheme = "http"
		req.URL.Host = upstream
		log.Printf("%s -> %s", req.Host, upstream)
	}
	return &http.Server{
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  3 * time.Minute,
		Handler:      &httputil.ReverseProxy{Director: director},
		Addr:         addr,
	}
}

func buildMapAndDefUpstream(mapList []string) (upstreamMap, regexpMap, string, error) {
	var defUpstream = ""
	var uMap, rMap = make(upstreamMap, len(mapList)), make(regexpMap)
	for _, mapping := range mapList {
		hostToUpstream := strings.SplitN(mapping, "@", 2)
		if len(hostToUpstream) != 2 {
			return uMap, rMap, defUpstream, fmt.Errorf("Wrong mapping: %s", mapping)
		}
		host, upstream := hostToUpstream[0], hostToUpstream[1]
		if strings.Contains(host, "*") {
			pattern := strings.Replace(regexp.QuoteMeta(host), `\*`, ".+", -1)
			rMap[regexp.MustCompile("^" + pattern + "$")] = upstream
		} else {
			uMap[host] = upstream
		}
		if defUpstream == "" {
			defUpstream = upstream
		}
	}
	return uMap, rMap, defUpstream, nil
}

func main() {
	var addr, cert, key string
	flag.StringVar(&addr, "addr", ":http", "which address listen to")
	flag.StringVar(&cert, "cert", "", "path to the certificate file (requires -key)")
	flag.StringVar(&key, "key", "", "path to the key file (requires -cert)")
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-addr=[iface]:port] [(-cert=cert -key=key)] domain@upstream [domain@upstream ...]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println("  domain@upstream")
		fmt.Println("    \tdomain may contain simple placeholders, such as *.domain.name")
	}
	flag.Parse()
	uMap, rMap, defUpstream, err := buildMapAndDefUpstream(flag.Args())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if len(uMap) == 0 || ((cert != "") != (key != "")) {
		flag.Usage()
		os.Exit(1)
	}
	proxy := buildProxy(addr, uMap, rMap, defUpstream)
	if cert == "" {
		log.Fatal(proxy.ListenAndServe())
	} else {
		log.Fatal(proxy.ListenAndServeTLS(cert, key))
	}
}
