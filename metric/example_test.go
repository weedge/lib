package metric

import (
	"github.com/gin-gonic/gin"
)

func ExampleHttpMetrics() {
	e := gin.Default()
	HttpMetrics(e)
	e.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	err := e.Run(":8081") // listen and serve on 0.0.0.0:8081 (for windows "localhost:8081")
	if err != nil {
		return
	}

	// output:
	//
}
