package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const urlPrefix = "/rest"

var servers = map[string]string{"http://localhost:3002": "/temp", "http://localhost:3004": "/other"}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

// Log the typeform payload and redirect url
func logRequestPayload(proxyURL string) {
	log.Printf("proxy_url: %s\n", proxyURL)
}

// Balance returns one of the servers based using round-robin algorithm
func getProxyURL(path string) (string, bool) {

	for server, route := range servers {
		if strings.HasPrefix(path, route) {
			return server, true
		}
	}
	return "", false
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {

	r.URL.Path = strings.TrimPrefix(r.URL.Path, urlPrefix)

	url, found := getProxyURL(r.URL.Path)

	if !found {
		return
	}

	logRequestPayload(url)

	serveReverseProxy(url, w, r)
}

func serverPaths(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(servers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	// start server
	http.HandleFunc(urlPrefix, serverPaths)

	http.HandleFunc(fmt.Sprint(urlPrefix, "/"), handleRequestAndRedirect)

	log.Fatal(http.ListenAndServe(":"+"3001", nil))
}
