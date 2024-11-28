package uam

import (
	jwts "core/app/common/jwt"
	"core/app/helper"
	gormDb "march-auth/cmd/app/common/gorm"
	"march-auth/cmd/app/graph/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DiviceId(c *gin.Context) {
	logctx := helper.LogContext("UAMSERVICE", "DiviceId")

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	tokenSplit := strings.Split(token, "Bearer ")
	logctx.Logger(tokenSplit, "tokenSplit")

	if len(tokenSplit) != 2 || strings.TrimSpace(tokenSplit[1]) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		c.Abort()
		return
	}

	actualToken := strings.TrimSpace(tokenSplit[1])

	validate, err := jwts.Verify(actualToken)

	if err != nil || !validate.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	userInfo, err := jwts.Decode(actualToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	user := &model.User{}
	gormDb.Repos.Where("id = ?", userInfo.UserId).Find(user)

	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deviceId": user.DeviceID,
	})

}
