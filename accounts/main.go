package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var assets embed.FS

func main() {

	r := gin.Default()

	fsys, _ := fs.Sub(assets, "dist")
	h := http.StripPrefix("", http.FileServer(http.FS(fsys)))
	r.GET("/*path", func(c *gin.Context) {
		fmt.Println(c.Request.URL.Path)
		h.ServeHTTP(c.Writer, c.Request)
	})

	log.Println(r.Run(":4200"))

}
