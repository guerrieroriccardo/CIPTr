package handlers

import "github.com/gin-gonic/gin"

// ok sends a successful JSON response: {"data": ..., "error": null}
func ok(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

// fail sends an error JSON response: {"data": null, "error": "..."}
func fail(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"data": nil, "error": err.Error()})
}
