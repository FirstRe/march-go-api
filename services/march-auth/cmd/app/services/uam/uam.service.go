package uam

import (
	"github.com/gin-gonic/gin"
)

func DiviceId(c *gin.Context) {
	c.JSON(200, gin.H{
		"deviceId": "bfcc2693-8db1-43be-9f5e-b6f91836da73",
	})
	// bfcc2693-8db1-43be-9f5e-b6f91836da73
	// c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token2"})
	// c.Abort()
	// return
}
