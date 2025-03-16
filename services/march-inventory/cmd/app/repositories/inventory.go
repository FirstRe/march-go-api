package repositories

import (
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
)

type InventoryRepository interface {
	GetInventories(searchParam string, isSerialNumber bool, params *types.ParamsInventory, shopId string) ([]model.Inventory, int, error)
	FindFirstInventory(where map[string]interface{}) (model.Inventory, error)
	UpdateInventory(id string, updatedData map[string]interface{}) error
	SaveInventory(inventoryData model.Inventory) error
}
