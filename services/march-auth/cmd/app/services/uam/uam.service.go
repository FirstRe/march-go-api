package uam

import (
	"github.com/gin-gonic/gin"
)

func DiviceId(c *gin.Context) {
	c.JSON(200, gin.H{
		"deviceId": "26f3fdb4-b475-496b-8d6b-0fc26b252d35",
	})
	// bfcc2693-8db1-43be-9f5e-b6f91836da73
	// c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token2"})
	// c.Abort()
	// return
}
