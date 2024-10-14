package middlewares

import (
	"bytes"
	"context"
	"core/app/common/jwt"
	"encoding/json"
	"io"
	"log"

	// "myapp/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type UserClaims struct {
	UserInfo      jwt.JwtCustomClaim
	Lang          string
	OperationName string
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type ErrorResponse struct {
	message string      `json:"message"`
	code    string      `json:"code"`
	Data    interface{} `json:"data,omitempty"`
}

func AuthString(key string) string {
	return "auth_" + key
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var gqlErr *gqlerror.Error
		auth := c.GetHeader("Authorization")
		lang := c.GetHeader("lang")

		if auth == "" {
			c.Next()
			return
		}

		bearer := "Bearer "
		if len(auth) > len(bearer) && auth[:len(bearer)] == bearer {
			auth = auth[len(bearer):] // Strip the "Bearer " prefix
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		validate, err := jwt.JwtValidate(context.Background(), auth)
		if err != nil || !validate.Valid {
			gqlErr = gqlerror.Errorf("Invalid Token or Authentication Failed")

		}

		userInfo, err := jwt.VerifyJWT(auth)
		if err != nil {
			log.Printf("err: %v", err)
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Token3"})
			c.Abort()
			return
		}

		var req GraphQLRequest

		if c.Request.Method == http.MethodPost && c.GetHeader("Content-Type") == "application/json" {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			_ = json.Unmarshal(body, &req)

		}

		ctx := context.WithValue(c.Request.Context(), AuthString("auth"), auth)
		ctx = context.WithValue(ctx, AuthString("userInfo"), userInfo)
		ctx = context.WithValue(ctx, AuthString("lang"), lang)
		if gqlErr != nil {
			ctx = context.WithValue(ctx, AuthString("gqlError"), gqlErr)
		}

		if req.OperationName != "" {
			log.Println("Operation Name:", req.OperationName)
			ctx = context.WithValue(ctx, AuthString("operationName"), req.OperationName)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func CtxValue(ctx context.Context) string {
	raw, _ := ctx.Value(AuthString("auth")).(string)
	
	return raw
}

func GqlErr(ctx context.Context) *gqlerror.Error {
	raw, _ := ctx.Value(AuthString("gqlError")).(*gqlerror.Error)
	return raw
}

func UserInfo(ctx context.Context) UserClaims {
	var claims UserClaims
	if userInfo, ok := ctx.Value(AuthString("userInfo")).(*jwt.JwtCustomClaim); ok {
		claims.UserInfo = *userInfo
	}
	if lang, ok := ctx.Value(AuthString("lang")).(string); ok {
		claims.Lang = lang
	}
	if operationName, ok := ctx.Value(AuthString("operationName")).(string); ok {
		claims.OperationName = operationName
	}

	return claims
}
