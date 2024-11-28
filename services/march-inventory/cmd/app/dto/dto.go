package dto

import (
	. "core/app/helper"
	"core/app/middlewares"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	"time"
)

func MapInputToInventory(input types.UpsertInventoryInput, userInfo middlewares.UserClaims) model.Inventory {
	var expiryDate *time.Time
	if input.ExpiryDate != nil {
		parsedDate, err := time.Parse(time.RFC3339, *input.ExpiryDate)
		if err == nil {
			expiryDate = &parsedDate
		}
	}

	return model.Inventory{
		ID:                DefaultTo(input.ID, ""),
		Name:              input.Name + "|" + userInfo.UserInfo.ShopsID,
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
		ShopsID:           userInfo.UserInfo.ShopsID,
		CreatedBy:         userInfo.UserInfo.UserName,
		UpdatedBy:         userInfo.UserInfo.UserName,
		Favorite:          DefaultTo(input.Favorite, false),
	}
}
