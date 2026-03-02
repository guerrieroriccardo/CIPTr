package handlers

import "github.com/gin-gonic/gin"

// ok sends a successful JSON response with the standard envelope format:
//
//	{"data": <payload>, "error": null}
//
// The status parameter is the HTTP status code (e.g. 200, 201).
func ok(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "error": nil})
}

// fail sends an error JSON response with the standard envelope format:
//
//	{"data": null, "error": "<message>"}
//
// The status parameter is the HTTP error code (e.g. 400, 404, 500).
func fail(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"data": nil, "error": err.Error()})
}
