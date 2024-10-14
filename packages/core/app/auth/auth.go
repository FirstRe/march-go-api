package auth

import (
	"context"
	"core/app/helper"
	"core/app/middlewares"

	// "log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	userInfo := middlewares.UserInfo(ctx)
	gqlErr := middlewares.GqlErr(ctx)
	l := helper.LogContext("AuthMiddleware", "Auth")

	l.Logger(userInfo, "userInfo", true)
	l.Logger(gqlErr, "gqlErr", true)

	if gqlErr != nil {
		return nil, &gqlerror.Error{
			Message: gqlErr.Message,
		}
	}

	if userInfo.UserInfo.ShopsID == "" {
		return nil, &gqlerror.Error{
			Message: "Access Denied",
		}
	}

	return next(ctx)
}
