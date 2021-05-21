package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var assets embed.FS

func main() {

	r := gin.Default()

	h := http.StripPrefix("", http.FileServer(http.FS(assets)))
	r.GET("/*path", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})

	log.Println(r.Run(":4200"))

}
