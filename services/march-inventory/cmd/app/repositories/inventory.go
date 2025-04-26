package repositories

import (
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
)

type InventoryRepository interface {
	GetInventories(searchParam string, isSerialNumber bool, params *types.ParamsInventory, shopId string) ([]model.Inventory, int, error)
	FindFirstInventory(params FindParams) (inventory model.Inventory, err error)
	FindInventory(params FindParams) (inventories []model.Inventory, err error)
	UpdateInventory(id string, updatedData map[string]interface{}) error
	SaveInventory(inventoryData model.Inventory) error

	//branch
	FindFirstInventoryBranch(params FindParams) (model.InventoryBranch, error)
	FindInventoryBranch(params FindParams) (inventoryBranchs []model.InventoryBranch, err error)

	//brand
	FindFirstInventoryBrand(params FindParams) (model.InventoryBrand, error)
	FindInventoryBrand(params FindParams) (inventoryBrands []model.InventoryBrand, err error)

	//type
	FindFirstInventoryType(params FindParams) (model.InventoryType, error)
	FindInventoryType(params FindParams) (inventoryTypes []model.InventoryType, err error)

	//sub
	DeleteSubInventory(checkIn interface{}, id string) error

	RecoverySubInventory(checkIn interface{}, id string, updatedData map[string]interface{}, updatedBy string) error
}
