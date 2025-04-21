package newInventory

import (
	"core/app/middlewares"
	"march-inventory/cmd/app/graph/types"
)

type InventoryService interface {
	GetInventories(params *types.ParamsInventory, userInfo middlewares.UserClaims) (*types.InventoriesResponse, error)
	FavoriteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	UpsertInventory(input types.UpsertInventoryInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	DeleteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	GetInventory(id *string, userInfo middlewares.UserClaims) (*types.InventoryDataResponse, error)
	DeleteInventoryCache(key string) error
}
