package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MinCLIVersion is the minimum CLI version accepted by this backend.
// Bump this when making breaking API changes.
const MinCLIVersion = "0.3.2"

// CLIVersionRequired rejects requests from CLI versions older than MinCLIVersion.
// Requests without the X-CLI-Version header (e.g. browser, curl) are allowed through.
func CLIVersionRequired() gin.HandlerFunc {
	minMajor, minMinor, minPatch := parseSemver(MinCLIVersion)

	return func(c *gin.Context) {
		v := c.GetHeader("X-CLI-Version")
		if v == "" || v == "dev" {
			c.Next()
			return
		}

		major, minor, patch := parseSemver(v)
		if compareSemver(major, minor, patch, minMajor, minMinor, minPatch) < 0 {
			c.AbortWithStatusJSON(http.StatusUpgradeRequired, gin.H{
				"data":  nil,
				"error": fmt.Sprintf("CLI version %s is too old, minimum required: %s. Run: ciptr-cli update", v, MinCLIVersion),
			})
			return
		}

		c.Next()
	}
}

// parseSemver extracts major, minor, patch from a "X.Y.Z" string.
// Returns (0, 0, 0) on any parse failure.
func parseSemver(s string) (int, int, int) {
	var major, minor, patch int
	fmt.Sscanf(s, "%d.%d.%d", &major, &minor, &patch)
	return major, minor, patch
}

// compareSemver returns -1, 0, or 1.
func compareSemver(aMajor, aMinor, aPatch, bMajor, bMinor, bPatch int) int {
	if aMajor != bMajor {
		return cmpInt(aMajor, bMajor)
	}
	if aMinor != bMinor {
		return cmpInt(aMinor, bMinor)
	}
	return cmpInt(aPatch, bPatch)
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
