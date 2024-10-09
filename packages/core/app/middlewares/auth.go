package middlewares

import (
	"context"
	"core/app/common/jwt"
	"log"

	// "myapp/service"
	"net/http"
)

type AuthString string

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			next.ServeHTTP(w, r)
			return
		}
		bearer := "Bearer "
		if len(auth) > len(bearer) && auth[:len(bearer)] == bearer {
			auth = auth[len(bearer):]
		} else {
			http.Error(w, "Invalid Token", http.StatusForbidden)
			return
		}

		validate, err := jwt.JwtValidate(context.Background(), auth)
		log.Printf("validate: %+v\n", validate)
		if err != nil || !validate.Valid {

			// http.Error(w, "Invalid token", http.StatusForbidden)
			next.ServeHTTP(w, r)
			return
		}
		userInfo, err := jwt.VerifyJWT(auth)
		if err != nil {
			log.Printf("err:%v", err)
			http.Error(w, "Invalid Token", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), AuthString("auth"), auth)
		ctx = context.WithValue(ctx, AuthString("userInfo"), userInfo)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func CtxValue(ctx context.Context) string {
	raw, _ := ctx.Value(AuthString("auth")).(string)
	return raw
}

func UserInfo(ctx context.Context) *jwt.JwtCustomClaim {
	raw, _ := ctx.Value(AuthString("userInfo")).(*jwt.JwtCustomClaim)
	return raw
}
