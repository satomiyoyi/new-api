package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		if uri == "/" || uri == "/index.html" {
			c.Header("Cache-Control", "no-cache")
		} else if strings.HasPrefix(uri, "/assets/") {
			// Vite hashed assets: cache immutable for 1 year
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "public, max-age=604800") // one week
		}
		c.Next()
	}
}
