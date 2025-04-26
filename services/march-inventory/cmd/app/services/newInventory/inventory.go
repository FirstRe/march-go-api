package newInventory

import (
	"core/app/middlewares"
	"march-inventory/cmd/app/graph/types"

	"github.com/99designs/gqlgen/graphql"
)

type InventoryService interface {
	GetInventories(params *types.ParamsInventory, userInfo middlewares.UserClaims) (*types.InventoriesResponse, error)
	FavoriteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	UpsertInventory(input types.UpsertInventoryInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	DeleteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	GetInventoryNames(userInfo middlewares.UserClaims) (*types.InventoryNameResponse, error)
	GetInventory(id *string, userInfo middlewares.UserClaims) (*types.InventoryDataResponse, error)
	GetInventoryAllDeleted(userInfo middlewares.UserClaims) (*types.DeletedInventoryResponse, error)
	RecoveryHardDeleted(input types.RecoveryHardDeletedInput, userInfo middlewares.UserClaims) (*types.RecoveryHardDeletedResponse, error)
	DeleteInventoryCache(key string) error
	UploadCsv(file graphql.Upload, userInfo middlewares.UserClaims) (*types.UploadInventoryResponse, error)

	//branch
	UpsertInventoryBranch(input *types.UpsertInventoryBranchInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error)
	DeleteInventoryBranch(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error)
	GetInventoryBranchs(params *types.ParamsInventoryBranch, userInfo middlewares.UserClaims) (*types.InventoryBranchsDataResponse, error)
	GetInventoryBranch(id *string) (*types.InventoryBranch, error)

	//type
	UpsertInventoryType(input *types.UpsertInventoryTypeInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	DeleteInventoryType(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error)
	GetInventoryTypes(params *types.ParamsInventoryType, userInfo middlewares.UserClaims) (*types.InventoryTypesResponse, error)
	GetInventoryType(id *string) (*types.InventoryTypeResponse, error)

	//brand
	UpsertInventoryBrand(input *types.UpsertInventoryBrandInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error)
	DeleteInventoryBrand(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error)
	GetInventoryBrands(params *types.ParamsInventoryBrand, userInfo middlewares.UserClaims) (*types.InventoryBrandsDataResponse, error)
	GetInventoryBrand(id *string) (*types.InventoryBrand, error)
}
