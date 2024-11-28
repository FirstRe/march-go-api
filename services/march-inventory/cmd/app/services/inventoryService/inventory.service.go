package inventoryService

import (
	"core/app/helper"
	. "core/app/helper"
	"core/app/middlewares"
	"errors"
	"log"
	"march-inventory/cmd/app/dto"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/statusCode"
	gormDb "march-inventory/cmd/app/statusCode/gorm"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertInventory(input types.UpsertInventoryInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventory")
	logctx.Logger(input, "input")
	findDup := model.Inventory{}
	name := input.Name + "|" + userInfo.UserInfo.ShopsID
	gormDb.Repos.Model(&model.Inventory{}).Where("name = ? AND shops_Id = ?", name, userInfo.UserInfo.ShopsID).First(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(findDup, "findDup")
	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	findDup = model.Inventory{
		ID:          "",
		Name:        name,
		Description: input.Description,
		ShopsID:     userInfo.UserInfo.ShopsID,
		CreatedBy:   userInfo.UserInfo.UserName,
		UpdatedBy:   userInfo.UserInfo.UserName,
	}
	inventoryData := dto.MapInputToInventory(input, userInfo)
	onOkLocalT := "Upsert.success.create.inventory"
	saveFailedLocalT := "Upsert.failed.create"

	if input.ID != nil {
		findDup.ID = *input.ID
		onOkLocalT = "Upsert.success.update.inventory"
		saveFailedLocalT = "Upsert.failed.update"
		if findDup.ShopsID != userInfo.UserInfo.ShopsID {
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.Forbidden("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
	}

	logctx.Logger(inventoryData, "InventoryData", true)

	if result := gormDb.Repos.Save(&inventoryData); result.Error != nil {
		logctx.Logger(result.Error, "[error-api] Upsert")
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError(translation.LocalizeMessage(saveFailedLocalT)),
			Data:   nil,
		}
		return &reponseError, nil
	} else {
		reponsePass := types.MutationInventoryResponse{
			Status: statusCode.Success(translation.LocalizeMessage(onOkLocalT)),
			Data: &types.ResponseID{
				ID: &findDup.ID,
			},
		}
		return &reponsePass, nil
	}

}

func DeleteInventoryTypes(id string) (*types.MutationInventoryResponse, error) {
	logctx := LogContext(ClassName, "DeleteInventoryType")
	logctx.Logger([]interface{}{id}, "id")

	inventoryType := model.InventoryType{}
	if err := gormDb.Repos.Model(&model.InventoryType{}).Where("id = ?", id).First(&inventoryType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("InventoryType with id %s not found", id)
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.NotFound("Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		log.Printf("Error fetching InventoryType: %+v", err)
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Internal server error"),
			Data:   nil,
		}
		return &reponseError, err
	}

	if err := gormDb.Repos.Model(&inventoryType).Update("deleted", true).Error; err != nil {
		log.Printf("Error deleting InventoryType: %+v", err)
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Error deleting inventory type"),
			Data:   nil,
		}
		return &reponseError, err
	}

	reponseSuccess := types.MutationInventoryResponse{
		Status: statusCode.Success("OK"),
		Data:   nil,
	}
	return &reponseSuccess, nil
}

func GetInventories(params *types.ParamsInventory, userInfo middlewares.UserClaims) (*types.InventoriesResponse, error) {
	logctx := LogContext(ClassName, "GetInventories")
	logctx.Logger(params, "params")

	pageNo := DefaultTo(params.PageNo, 1)
	limit := DefaultTo(params.Limit, 30)
	offset := pageNo*limit - limit

	logctx.Logger(offset, "offset")
	logctx.Logger(pageNo, "pageNo")
	logctx.Logger(limit, "limit")

	inventories := []model.Inventory{}

	searchParam := ""
	isSerialNumber := false
	if params != nil && params.Search != nil {
		if strings.HasPrefix(*params.Search, "#") {
			isSerialNumber = true
		}
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	query := gormDb.Repos.Model(&[]model.Inventory{}).Where("deleted = ?", false)

	if searchParam != "" && !isSerialNumber {
		query = query.Where("name LIKE ?", searchParam)
	} else if isSerialNumber {
		query = query.Where("serial_number LIKE ?", searchParam)
	}

	if params.Favorite != nil && *params.Favorite == types.FavoriteStatusLike {
		query = query.Where("favorite = ?", true)
	}

	if len(params.Type) > 0 {
		query = query.Where("inventory_type_id IN ?", params.Type).Preload("InventoryType")
	}

	if len(params.Brand) > 0 {
		query = query.Where("inventory_brand_id IN ?", params.Brand).Preload("InventoryBrand")
	}

	if len(params.Branch) > 0 {
		query = query.Where("inventory_branch_id IN ?", params.Branch).Preload("InventoryBranch")
	}

	if userInfo.UserInfo.ShopsID == "" || userInfo.UserInfo.UserName == "" {
		reponseError := types.InventoriesResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	query = query.Where("shops_id = ?", userInfo.UserInfo.ShopsID)

	result := query.Preload(clause.Associations).Order("created_at desc").Limit(limit).Offset(offset).Find(&inventories)

	if result.Error != nil {
		logctx.Logger([]interface{}{result.Error}, "err GetInventories Model Data")
		reponseError := types.InventoriesResponse{
			Status: statusCode.InternalError(""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(inventories, "[log-api] inventories")

	inventoriesData := make([]*types.Inventory, len(inventories))

	for d, inventory := range inventories {
		inventoryBrand := types.InventoryBrand{
			ID:          &inventory.InventoryBrand.ID,
			Name:        inventory.InventoryBrand.Name,
			Description: inventory.InventoryBrand.Description,
			CreatedBy:   &inventory.InventoryBrand.CreatedBy,
			CreatedAt:   inventory.InventoryBrand.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryBrand.UpdatedBy,
			UpdatedAt:   inventory.InventoryBrand.UpdatedAt.String(),
		}

		inventoryBranch := types.InventoryBranch{
			ID:          &inventory.InventoryBranch.ID,
			Name:        inventory.InventoryBranch.Name,
			Description: inventory.InventoryBranch.Description,
			CreatedBy:   &inventory.InventoryBranch.CreatedBy,
			CreatedAt:   inventory.InventoryBranch.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryBranch.UpdatedBy,
			UpdatedAt:   inventory.InventoryBranch.UpdatedAt.String(),
		}

		inventoryType := types.InventoryType{
			ID:          &inventory.InventoryType.ID,
			Name:        inventory.InventoryType.Name,
			Description: inventory.InventoryType.Description,
			CreatedBy:   &inventory.InventoryType.CreatedBy,
			CreatedAt:   inventory.InventoryType.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryType.UpdatedBy,
			UpdatedAt:   inventory.InventoryType.UpdatedAt.String(),
		}

		expiryDateStr := ""
		if inventory.ExpiryDate != nil {
			expiryDateStr = inventory.ExpiryDate.String()
		}

		inventoriesData[d] = &types.Inventory{
			ID:              &inventory.ID,
			Name:            inventory.Name,
			Description:     inventory.Description,
			CreatedBy:       &inventory.CreatedBy,
			CreatedAt:       inventory.CreatedAt.String(),
			UpdatedBy:       &inventory.UpdatedBy,
			UpdatedAt:       inventory.UpdatedAt.String(),
			Amount:          inventory.Amount,
			Sold:            &inventory.Sold,
			Sku:             inventory.SKU,
			SerialNumber:    inventory.SerialNumber,
			Size:            inventory.Size,
			PriceMember:     inventory.PriceMember,
			Price:           inventory.Price,
			ReorderLevel:    inventory.ReorderLevel,
			ExpiryDate:      &expiryDateStr,
			InventoryBrand:  &inventoryBrand,
			InventoryBranch: &inventoryBranch,
			InventoryType:   &inventoryType,
			Favorite:        &inventory.Favorite,
		}
	}
	logctx.Logger(inventories, "[log-api] inventories1")
	var count int64
	if err := query.Find(&inventories).Count(&count).Error; err != nil {
		count = 0
	}
	totalRow := int(count)
	totalPage := int(math.Ceil(float64(totalRow) / float64(limit)))

	reponseData := types.ResponseInventories{
		Inventories: inventoriesData,
		PageLimit:   params.Limit,
		TotalRow:    &totalRow,
		TotalPage:   &totalPage,
		PageNo:      params.PageNo,
	}

	reponsePass := types.InventoriesResponse{
		Status: statusCode.Success("OK"),
		Data:   &reponseData,
	}

	return &reponsePass, nil
}

func GetInventoryNames(userInfo middlewares.UserClaims) (*types.InventoryNameResponse, error) {
	logctx := LogContext(ClassName, "GetInventories")

	inventories := []model.Inventory{}
	gormDb.Repos.Where("shops_id = ?", userInfo.UserInfo.ShopsID).Find(&inventories).Select("id", "name")

	logctx.Logger(inventories, "inventories")
	inventoryName := make([]*types.InventoryName, len(inventories))

	for d, inventory := range inventories {
		inventoryName[d] = &types.InventoryName{
			ID:   &inventory.ID,
			Name: &inventory.Name,
		}
	}

	reponseSuccess := types.InventoryNameResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryName,
	}
	return &reponseSuccess, nil

}

func GetInventoryAllDeleted(userInfo middlewares.UserClaims) (*types.DeletedInventoryResponse, error) {
	logctx := LogContext(ClassName, "GetInventoryAllDeleted")
	logctx.Logger(userInfo, "userInfo")

	inventories := []model.Inventory{}
	gormDb.Repos.
		Where("shops_id = ? AND deleted = ?", userInfo.UserInfo.ShopsID, true).
		Find(&inventories).
		Select("id", "name").
		Order("updated_at desc")

	logctx.Logger(inventories, "inventories")
	deletedInventory := make([]*types.DeletedInventoryType, len(inventories))

	for d, inventory := range inventories {
		deletedInventory[d] = &types.DeletedInventoryType{
			ID:        &inventory.ID,
			Name:      &inventory.Name,
			CreatedBy: &inventory.CreatedBy,
			UpdatedBy: &inventory.UpdatedBy,
			UpdatedAt: inventory.UpdatedAt.String(),
			CreatedAt: inventory.CreatedAt.String(),
		}
	}

	inventoryTypes := []model.InventoryType{}
	gormDb.Repos.
		Where("shops_id = ? AND deleted = ?", userInfo.UserInfo.ShopsID, true).
		Find(&inventoryTypes).
		Select("id", "name").
		Order("updated_at desc")

	logctx.Logger(inventoryTypes, "inventoryTypes")
	deletedInventoryType := make([]*types.DeletedInventoryType, len(inventoryTypes))

	for d, inventoryType := range inventoryTypes {
		deletedInventoryType[d] = &types.DeletedInventoryType{
			ID:        &inventoryType.ID,
			Name:      &inventoryType.Name,
			CreatedBy: &inventoryType.CreatedBy,
			UpdatedBy: &inventoryType.UpdatedBy,
			UpdatedAt: inventoryType.UpdatedAt.String(),
			CreatedAt: inventoryType.CreatedAt.String(),
		}
	}

	inventoryBrands := []model.InventoryBrand{}
	gormDb.Repos.
		Where("shops_id = ? AND deleted = ?", userInfo.UserInfo.ShopsID, true).
		Find(&inventoryBrands).
		Select("id", "name").
		Order("updated_at desc")

	logctx.Logger(inventoryBrands, "inventoryBrands")
	deletedInventoryBrand := make([]*types.DeletedInventoryType, len(inventoryBrands))

	for d, inventoryBrand := range inventoryBrands {
		deletedInventoryBrand[d] = &types.DeletedInventoryType{
			ID:        &inventoryBrand.ID,
			Name:      &inventoryBrand.Name,
			CreatedBy: &inventoryBrand.CreatedBy,
			UpdatedBy: &inventoryBrand.UpdatedBy,
			UpdatedAt: inventoryBrand.UpdatedAt.String(),
			CreatedAt: inventoryBrand.CreatedAt.String(),
		}
	}

	inventoryBranchs := []model.InventoryBranch{}
	gormDb.Repos.
		Where("shops_id = ? AND deleted = ?", userInfo.UserInfo.ShopsID, true).
		Find(&inventoryBranchs).
		Select("id", "name").
		Order("updated_at desc")

	logctx.Logger(inventoryBranchs, "inventoryBranchs")
	deletedInventoryBranch := make([]*types.DeletedInventoryType, len(inventoryBranchs))

	for d, inventoryBranch := range inventoryBranchs {
		deletedInventoryBranch[d] = &types.DeletedInventoryType{
			ID:        &inventoryBranch.ID,
			Name:      &inventoryBranch.Name,
			CreatedBy: &inventoryBranch.CreatedBy,
			UpdatedBy: &inventoryBranch.UpdatedBy,
			UpdatedAt: inventoryBranch.UpdatedAt.String(),
			CreatedAt: inventoryBranch.CreatedAt.String(),
		}
	}

	response := types.DeletedInventoryResponse{
		Data: &types.DeletedInventory{
			Inventory: deletedInventory,
			Type:      deletedInventoryType,
			Brand:     deletedInventoryBrand,
			Branch:    deletedInventoryBranch,
		},
		Status: statusCode.Success("OK"),
	}

	return &response, nil
}

func GetInventory(id *string, userInfo middlewares.UserClaims) (*types.InventoryDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventory")
	logctx.Logger(id, "id")
	inventory := &model.Inventory{}
	gormDb.Repos.Where("id = ?", id).
		Preload("InventoryType").
		Preload("InventoryBranch").
		Preload("InventoryBrand").
		First(inventory)
	logctx.Logger(inventory, "inventory")

	if userInfo.UserInfo.ShopsID == "" || userInfo.UserInfo.UserName == "" {
		reponseError := types.InventoryDataResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if inventory.ShopsID != userInfo.UserInfo.ShopsID {
		reponseError := types.InventoryDataResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventoryBrand := types.InventoryBrand{
		ID:          &inventory.InventoryBrand.ID,
		Name:        inventory.InventoryBrand.Name,
		Description: inventory.InventoryBrand.Description,
		CreatedBy:   &inventory.InventoryBrand.CreatedBy,
		CreatedAt:   inventory.InventoryBrand.CreatedAt.String(),
		UpdatedBy:   &inventory.InventoryBrand.UpdatedBy,
		UpdatedAt:   inventory.InventoryBrand.UpdatedAt.String(),
	}

	inventoryBranch := types.InventoryBranch{
		ID:          &inventory.InventoryBranch.ID,
		Name:        inventory.InventoryBranch.Name,
		Description: inventory.InventoryBranch.Description,
		CreatedBy:   &inventory.InventoryBranch.CreatedBy,
		CreatedAt:   inventory.InventoryBranch.CreatedAt.String(),
		UpdatedBy:   &inventory.InventoryBranch.UpdatedBy,
		UpdatedAt:   inventory.InventoryBranch.UpdatedAt.String(),
	}

	inventoryType := types.InventoryType{
		ID:          &inventory.InventoryType.ID,
		Name:        inventory.InventoryType.Name,
		Description: inventory.InventoryType.Description,
		CreatedBy:   &inventory.InventoryType.CreatedBy,
		CreatedAt:   inventory.InventoryType.CreatedAt.String(),
		UpdatedBy:   &inventory.InventoryType.UpdatedBy,
		UpdatedAt:   inventory.InventoryType.UpdatedAt.String(),
	}

	expiryDateStr := ""
	if inventory.ExpiryDate != nil {
		expiryDateStr = inventory.ExpiryDate.String()
	}

	inventoryData := &types.Inventory{
		ID:              &inventory.ID,
		Name:            inventory.Name,
		Description:     inventory.Description,
		CreatedBy:       &inventory.CreatedBy,
		CreatedAt:       inventory.CreatedAt.String(),
		UpdatedBy:       &inventory.UpdatedBy,
		UpdatedAt:       inventory.UpdatedAt.String(),
		Amount:          inventory.Amount,
		Sold:            &inventory.Sold,
		Sku:             inventory.SKU,
		SerialNumber:    inventory.SerialNumber,
		Size:            inventory.Size,
		PriceMember:     inventory.PriceMember,
		Price:           inventory.Price,
		ReorderLevel:    inventory.ReorderLevel,
		ExpiryDate:      &expiryDateStr,
		InventoryBrand:  &inventoryBrand,
		InventoryBranch: &inventoryBranch,
		InventoryType:   &inventoryType,
		Favorite:        &inventory.Favorite,
	}

	reponse := &types.InventoryDataResponse{
		Data:   inventoryData,
		Status: statusCode.Success("OK"),
	}

	return reponse, nil
}

func DeleteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventory")
	logctx.Logger(id, "id")

	inventory := &model.Inventory{}
	gormDb.Repos.Where("id = ?", id).First(inventory)

	if inventory.ShopsID != userInfo.UserInfo.ShopsID || inventory.ID == "" {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if err := gormDb.Repos.Model(&inventory).Update("deleted", true).Error; err != nil {
		logctx.Logger(id, "[error-api] deleting Inventory")
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Error deleting inventory inventory"),
			Data:   nil,
		}
		return &reponseError, err
	} else {
		reponseSuccess := types.MutationInventoryResponse{
			Status: statusCode.Success(translation.LocalizeMessage("Success.inventory")),
			Data:   nil,
		}
		return &reponseSuccess, nil
	}

}
