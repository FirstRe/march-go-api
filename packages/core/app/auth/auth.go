package auth

import (
	"context"
	"core/app/middlewares"

	// "myapp/middlewares"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {

	userInfo := middlewares.UserInfo(ctx)
	// log.Printf("userInfow %+v", userInfo)
	if &userInfo == nil {
		return nil, &gqlerror.Error{
			Message: "Access Denied",
		}
	}

	return next(ctx)
}
