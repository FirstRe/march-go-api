package resolvers

import model "go-graphql/cmd/app/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Users []*model.User
	Posts []*model.Post
}
