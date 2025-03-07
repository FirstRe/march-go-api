package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"core/app/helper"
	"march-auth/cmd/app/graph/types"
	authService "march-auth/cmd/app/services/auth"
)

// TokenExpire is the resolver for the tokenExpire field.
func (r *mutationResolver) TokenExpire(ctx context.Context, refreshToken string) (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "TokenExpire")
	logctx.Logger([]interface{}{}, "TokenExpire")
	return authService.TokenExpire(refreshToken)
}

// SignOut is the resolver for the signOut field.
func (r *mutationResolver) SignOut(ctx context.Context, id string) (*types.SignOutResponse, error) {
	logctx := helper.LogContext(ClassNameAuth, "SignOut")
	logctx.Logger([]interface{}{}, "SignOut")
	return authService.SignOut(id)
}

// VerifyAccessToken is the resolver for the verifyAccessToken field.
func (r *mutationResolver) VerifyAccessToken(ctx context.Context, token string) (*types.VerifyAccessTokenResponse, error) {
	logctx := helper.LogContext(ClassNameAuth, "VerifyAccessToken")
	logctx.Logger([]interface{}{}, "token")
	return authService.VerifyAccessToken(token)
}

// SignInOAuth is the resolver for the signInOAuth field.
func (r *mutationResolver) SignInOAuth(ctx context.Context, code string) (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "SignInOAuth")
	logctx.Logger(code, "code")
	return authService.SignInOAuth(code)
}

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
const ClassNameAuth string = "AuthResolver"
