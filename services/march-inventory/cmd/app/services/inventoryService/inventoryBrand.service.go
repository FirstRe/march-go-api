package inventoryService

import (
	"core/app/helper"
	"core/app/middlewares"
	"errors"
	"log"
	gormDb "march-inventory/cmd/app/common/gorm"
	"march-inventory/cmd/app/common/statusCode"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	translation "march-inventory/cmd/app/i18n"

	"gorm.io/gorm"
)

func UpsertInventoryBrand(input *types.UpsertInventoryBrandInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBrand")
	logctx.Logger(input, "input")
	findDup := model.InventoryBrand{}
	typeName := input.Name + "|" + userInfo.UserInfo.ShopsID
	gormDb.Repos.Model(&model.InventoryBrand{}).Where("name = ? AND shops_Id = ?", typeName, userInfo.UserInfo.ShopsID).First(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(findDup, "findDup")
	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.BadRequest("Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	findDup = model.InventoryBrand{
		ID:          "",
		Name:        typeName,
		Description: input.Description,
		ShopsID:     userInfo.UserInfo.ShopsID,
		CreatedBy:   userInfo.UserInfo.UserName,
		UpdatedBy:   userInfo.UserInfo.UserName,
	}

	onOkLocalT := "Upsert.success.create.brand"
	saveFailedLocalT := "Upsert.failed.create"

	if input.ID != nil {
		findDup.ID = *input.ID
		onOkLocalT = "Upsert.success.update.brand"
		saveFailedLocalT = "Upsert.failed.update"
		if findDup.ShopsID != userInfo.UserInfo.ShopsID {
			reponseError := types.MutationInventoryBrandResponse{
				Status: statusCode.Forbidden("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
	}

	logctx.Logger(findDup, "InventoryBrandData", true)

	if result := gormDb.Repos.Save(&findDup); result.Error != nil {
		logctx.Logger(result.Error, "[error-api] Upsert")
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.InternalError(translation.LocalizeMessage(saveFailedLocalT)),
			Data:   nil,
		}
		return &reponseError, nil
	} else {
		reponsePass := types.MutationInventoryBrandResponse{
			Status: statusCode.Success(translation.LocalizeMessage(onOkLocalT)),
			Data: &types.ResponseID{
				ID: &findDup.ID,
			},
		}
		return &reponsePass, nil
	}

}

func DeleteInventoryBrand(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBrand")
	logctx.Logger(id, "id")

	inventoryBrand := model.InventoryBrand{}
	if err := gormDb.Repos.Where("id = ?", id).First(&inventoryBrand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger(id, "[error-api] InventoryBrand with id not found")
			reponseError := types.MutationInventoryBrandResponse{
				Status: statusCode.BadRequest("Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		logctx.Logger(err.Error(), "[error-api] fetching InventoryBrand")

		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.InternalError(""),
			Data:   nil,
		}
		return &reponseError, err
	}

	if inventoryBrand.ShopsID != userInfo.UserInfo.ShopsID {
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventory := model.Inventory{}
	result := gormDb.Repos.Where("inventory_brand_id = ?", id).First(&inventory)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryBrand not found")

		if err := gormDb.Repos.Model(&inventoryBrand).Update("deleted", true).Error; err != nil {
			logctx.Logger(id, "[error-api] deleting InventoryBrand")
			reponseError := types.MutationInventoryBrandResponse{
				Status: statusCode.InternalError("Error deleting inventory brand"),
				Data:   nil,
			}
			return &reponseError, err
		} else {
			reponseSuccess := types.MutationInventoryBrandResponse{
				Status: statusCode.Success(translation.LocalizeMessage("Success.brand")),
				Data:   nil,
			}
			return &reponseSuccess, nil
		}
	} else {
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("onUse.brand")),
			Data:   nil,
		}
		return &reponseError, nil
	}

}

func GetInventoryBrands(params *types.ParamsInventoryBrand, userInfo middlewares.UserClaims) (*types.InventoryBrandsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBrands")
	logctx.Logger([]interface{}{}, "id")

	inventoryBrands := []model.InventoryBrand{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if err := gormDb.Repos.Model(&inventoryBrands).Where("name LIKE ?", searchParam).Not("deleted = ?", true).Order("created_at asc").Find(&inventoryBrands).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryBrands Model Data")
	}
	logctx.Logger([]interface{}{inventoryBrands}, "inventoryBrands")
	inventoryBrandsData := make([]*types.InventoryBrand, len(inventoryBrands))

	for d, inventoryBrand := range inventoryBrands {
		inventoryBrandsData[d] = &types.InventoryBrand{
			ID:          &inventoryBrand.ID,
			Name:        inventoryBrand.Name,
			Description: inventoryBrand.Description,
			CreatedBy:   &inventoryBrand.CreatedBy,
			CreatedAt:   inventoryBrand.CreatedAt.String(),
			UpdatedBy:   &inventoryBrand.UpdatedBy,
			UpdatedAt:   inventoryBrand.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryBrandsDataResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryBrandsData,
	}

	return &reponsePass, nil
}

func GetInventoryBrand(id *string) (*types.InventoryBrand, error) {
	//notuse
	logctx := helper.LogContext(ClassName, "GetInventoryBrand")
	logctx.Logger([]interface{}{id}, "id")
	inventoryBrand := model.InventoryBrand{}
	if err := gormDb.Repos.Model(&inventoryBrand).Where("id = ?", id).First(&inventoryBrand).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryType Model Data")

		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger([]interface{}{}, "not found")
		}
		return nil, nil
	}
	logctx.Logger([]interface{}{inventoryBrand}, "InventoryBrand")

	inventoryBrandData := types.InventoryBrand{
		ID:          &inventoryBrand.ID,
		Name:        inventoryBrand.Name,
		Description: inventoryBrand.Description,
		CreatedBy:   &inventoryBrand.CreatedBy,
		CreatedAt:   inventoryBrand.CreatedAt.String(),
		UpdatedBy:   &inventoryBrand.UpdatedBy,
		UpdatedAt:   inventoryBrand.UpdatedAt.String(),
	}

	return &inventoryBrandData, nil
}
