package http

import "github.com/gin-gonic/gin"

type Routes interface {
	Register(*gin.Engine)
}
