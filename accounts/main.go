package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var assets embed.FS

func main() {
	proxy()
}

func fileserver() {

	r := gin.Default()

	fsys, _ := fs.Sub(assets, "dist")
	h := http.StripPrefix("", http.FileServer(http.FS(fsys)))
	r.GET("/*path", func(c *gin.Context) {
		fmt.Println(c.Request.URL.Path)
		h.ServeHTTP(c.Writer, c.Request)
	})

	log.Println(r.Run(":4200"))

}

func proxy() {
	origin, _ := url.Parse("http://localhost:4200/")

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
		req.URL.Path = "/proxy/9200" + req.URL.Path
		fmt.Println("director path : " + req.URL.Path)
		fmt.Println()
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("vscode   path : " + r.RequestURI)
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":9200", nil))
}
