package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const urlPrefix = "/rest"

var servers = map[string]string{"http://localhost:3002": "/temp", "http://localhost:3004": "/other"}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, c *gin.Context) {
	// parse the url
	remote, err := url.Parse(target)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(remote)

	logRequestPayload(remote.Host, c.Param("proxyPath"))

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("proxyPath")
	}

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(c.Writer, c.Request)
}

// Log the typeform payload and redirect url
func logRequestPayload(host, path string) {
	log.Printf("proxy_url: %s%s\n", host, path)
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
func proxy(c *gin.Context) {

	url, found := getProxyURL(c.Param("proxyPath"))

	if !found {
		return
	}

	serveReverseProxy(url, c)
}

func serverPaths(c *gin.Context) {
	c.JSON(http.StatusOK, servers)
}

func main() {

	r := gin.Default()

	r.GET(urlPrefix, serverPaths)

	//Create a catchall route
	r.Any(fmt.Sprint(urlPrefix, "/*proxyPath"), proxy)

	r.Run()
}
