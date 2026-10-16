package latihanGin

import (

	"github.com/gin-gonic/gin"
)

func Ping (c *gin.Context) {
	number := c.Query("number")
	c.JSON(200, gin.H{"Message": "Pong x " + number })
}

func Hello(c *gin.Context) {
	name := c.Param("name")
	c.JSON(200, gin.H{"Hello": "Hi too " + name})
}

func Run() {
	r := gin.Default()

	r.GET("/ping", Ping)
	r.GET("/Hello/:name", Hello)

	r.Run(":8080")

}
