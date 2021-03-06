package http

import (
	"github.com/go-kit/kit/endpoint"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthorizeMiddleware(authenticateEndpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/health-check" {
			c.Next()

			return
		}

		request := isAccessAllowedRequest{
			Header:   findToken(c),
			Resource: c.Request.RequestURI,
			Action:   c.Request.Method,
		}

		resp, err := authenticateEndpoint(c.Request.Context(), request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})

			return
		}

		response := resp.(isAccessAllowedResponse)

		if !response.Ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Permission denied",
			})

			return
		}

		c.Next()
	}
}

func findToken(c *gin.Context) string {
	if token := c.Request.Header.Get("Authorization"); token != "" {
		return token
	}

	v := c.Request.URL.Query()
	if token, exist := v["token"]; exist {
		return token[0]
	}

	return ""
}
