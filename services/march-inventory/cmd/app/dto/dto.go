package dto

import (
	. "core/app/helper"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	"time"
)

func MapInputToInventory(input types.UpsertInventoryInput) model.Inventory {
	var expiryDate *time.Time
	if input.ExpiryDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, input.ExpiryDate)
		if err == nil {
			expiryDate = &parsedDate
		}
	}

	return model.Inventory{
		ID:                DefaultTo(input.ID, ""),
		Name:              input.Name,
		InventoryTypeID:   input.InventoryTypeID,
		InventoryBrandID:  input.InventoryBrandID,
		InventoryBranchID: input.InventoryBranchID,
		Amount:            input.Amount,
		Price:             input.Price,
		PriceMember:       input.PriceMember,
		Size:              GetOptionalSize(input.Size),
		SKU:               input.Sku,
		SerialNumber:      input.SerialNumber,
		ReorderLevel:      input.ReorderLevel,
		ExpiryDate:        expiryDate,
		Description:       input.Description,
		CreatedBy:         DefaultTo(input.CreatedBy, ""),
		UpdatedBy:         DefaultTo(input.UpdatedBy, ""),
		Favorite:          DefaultTo(input.Favorite, false),
	}
}
