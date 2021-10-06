package tapper

import "github.com/gin-gonic/gin"

type ITraceLog interface {
	SetTraceLogFromGinHeader(c *gin.Context) *TraceLog
}

var TraceLogger ITraceLog
