package resolvers

import (
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/services/newInventory"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	InventoryBrands   []*model.InventoryBrand
	InventoryTypes    []*model.InventoryType
	InventoryBranches []*model.InventoryBranch
	Inventories       []*model.Inventory
	InventoryFiles    []*model.InventoryFile
	InventoryService  newInventory.InventoryService
}
