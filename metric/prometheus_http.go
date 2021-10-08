package metric

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func HttpMetrics(handler http.Handler) {
	switch r := handler.(type) {
	case *gin.Engine:
		h := promhttp.Handler()
		r.GET("/metrics", func(c *gin.Context) {
			h.ServeHTTP(c.Writer, c.Request)
		})
	}

}
