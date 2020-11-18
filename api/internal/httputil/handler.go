package httputil

import (
	"github.com/gin-gonic/gin"
)

// txid, logging, cors header, payload
type RequestHandler struct {
	Context *gin.Context
	Errors  []error
}

func NewRequestHandler(c *gin.Context) RequestHandler {
	return RequestHandler{
		c,
		[]error{},
	}
}
