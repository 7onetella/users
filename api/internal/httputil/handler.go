package httputil

import (
	"github.com/gin-gonic/gin"
)

// txid, logging, cors header, payload
type RequestHanlder struct {
	Context *gin.Context
	Errors  []error
}

func NewRequestHandler(c *gin.Context) RequestHanlder {
	return RequestHanlder{
		c,
		[]error{},
	}
}
