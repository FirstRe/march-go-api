package middlewares

import (
	"context"
	"core/app/common/jwt"
	"log"

	// "myapp/service"
	"net/http"
)

type UserClaims struct {
	UserInfo jwt.JwtCustomClaim
	Lang     string
}

func AuthString(key string) string {
	return "auth_" + key
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		lang := r.Header.Get("lang")

		if auth == "" {
			// Allow the request to continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		bearer := "Bearer "
		if len(auth) > len(bearer) && auth[:len(bearer)] == bearer {
			auth = auth[len(bearer):] // Strip the "Bearer " prefix
		} else {
			http.Error(w, "Invalid Token", http.StatusForbidden)
			return
		}

		// Validate the JWT
		validate, err := jwt.JwtValidate(context.Background(), auth)
		log.Printf("validate: %+v\n", validate)
		if err != nil || !validate.Valid {
			http.Error(w, "Invalid Token", http.StatusForbidden)
			return
		}

		// Verify and extract user info from the token
		userInfo, err := jwt.VerifyJWT(auth)
		if err != nil {
			log.Printf("err:%v", err)
			http.Error(w, "Invalid Token", http.StatusForbidden)
			return
		}

		// Set values in the context
		ctx := context.WithValue(r.Context(), AuthString("auth"), auth)
		ctx = context.WithValue(ctx, AuthString("userInfo"), userInfo)
		ctx = context.WithValue(ctx, AuthString("lang"), lang)

		// Create a new request with the updated context
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func CtxValue(ctx context.Context) string {
	raw, _ := ctx.Value(AuthString("auth")).(string)
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

	return claims
}
