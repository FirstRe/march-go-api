package inventoryService

import (
	"core/app/helper"
	"core/app/middlewares"
	"errors"
	"log"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/statusCode"
	gormDb "march-inventory/cmd/app/statusCode/gorm"

	"gorm.io/gorm"
)

const ClassName string = "InventoryService"

func UpsertInventoryType(input *types.UpsertInventoryTypeInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryType")
	logctx.Logger(input, "input")
	findDup := model.InventoryType{}
	typeName := input.Name + "|" + userInfo.UserInfo.ShopsID
	gormDb.Repos.Model(&model.InventoryType{}).Where("name = ? AND shops_Id = ?", typeName, userInfo.UserInfo.ShopsID).First(&findDup)

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

	findDup = model.InventoryType{
		ID:          "",
		Name:        typeName,
		Description: input.Description,
		ShopsID:     userInfo.UserInfo.ShopsID,
		CreatedBy:   userInfo.UserInfo.UserName,
		UpdatedBy:   userInfo.UserInfo.UserName,
	}

	onOkLocalT := "Upsert.success.create.type"
	saveFailedLocalT := "Upsert.failed.create"

	if input.ID != nil {
		findDup.ID = *input.ID
		onOkLocalT = "Upsert.success.update.type"
		saveFailedLocalT = "Upsert.failed.update"
		if findDup.ShopsID != userInfo.UserInfo.ShopsID {
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.Forbidden("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
	}

	logctx.Logger(findDup, "inventoryTypeData", true)

	if result := gormDb.Repos.Save(&findDup); result.Error != nil {
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

func DeleteInventoryType(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryType")
	logctx.Logger(id, "id")

	inventoryType := model.InventoryType{}
	if err := gormDb.Repos.Where("id = ?", id).First(&inventoryType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger(id, "[error-api] InventoryType with id not found")
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.BadRequest("Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		logctx.Logger(err, "[error-api] fetching InventoryType")

		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError(""),
			Data:   nil,
		}
		return &reponseError, err
	}

	if inventoryType.ShopsID != userInfo.UserInfo.ShopsID {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventory := model.Inventory{}
	result := gormDb.Repos.Where("inventory_type_id = ?", id).First(&inventory)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryType not found")

		if err := gormDb.Repos.Model(&inventoryType).Update("deleted", true).Error; err != nil {
			logctx.Logger(id, "[error-api] deleting InventoryType")
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.InternalError("Error deleting inventory type"),
				Data:   nil,
			}
			return &reponseError, err
		} else {
			reponseSuccess := types.MutationInventoryResponse{
				Status: statusCode.Success(translation.LocalizeMessage("Success.type")),
				Data:   nil,
			}
			return &reponseSuccess, nil
		}
	} else {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("onUse.type")),
			Data:   nil,
		}
		return &reponseError, nil
	}

}

func GetInventoryTypes(params *types.ParamsInventoryType, userInfo middlewares.UserClaims) (*types.InventoryTypesResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryTypes")
	logctx.Logger([]interface{}{}, "id")

	inventoryTypes := []model.InventoryType{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if err := gormDb.Repos.Model(&inventoryTypes).Where("name LIKE ?", searchParam).Not("deleted = ?", true).Order("created_at asc").Find(&inventoryTypes).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryTypes Model Data")
	}
	logctx.Logger([]interface{}{inventoryTypes}, "inventoryTypes")
	inventoryTypesData := make([]*types.InventoryType, len(inventoryTypes))

	for d, inventoryType := range inventoryTypes {
		inventoryTypesData[d] = &types.InventoryType{
			ID:          &inventoryType.ID,
			Name:        inventoryType.Name,
			Description: inventoryType.Description,
			CreatedBy:   &inventoryType.CreatedBy,
			CreatedAt:   inventoryType.CreatedAt.String(),
			UpdatedBy:   &inventoryType.UpdatedBy,
			UpdatedAt:   inventoryType.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryTypesResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryTypesData,
	}

	return &reponsePass, nil
}

func GetInventoryType(id *string) (*types.InventoryTypeResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryType")
	logctx.Logger([]interface{}{id}, "id")
	inventoryType := model.InventoryType{}
	if err := gormDb.Repos.Model(&inventoryType).Where("id = ?", id).First(&inventoryType).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryType Model Data")
		message := "Internal Error"

		if errors.Is(err, gorm.ErrRecordNotFound) {
			message = "Not Found"
		}

		reponseError := types.InventoryTypeResponse{
			Status: statusCode.BadRequest(message),
			Data:   nil,
		}
		return &reponseError, nil
	}
	logctx.Logger([]interface{}{inventoryType}, "inventoryType")

	inventoryTypeData := types.InventoryType{
		ID:          &inventoryType.ID,
		Name:        inventoryType.Name,
		Description: inventoryType.Description,
		CreatedBy:   &inventoryType.CreatedBy,
		CreatedAt:   inventoryType.CreatedAt.String(),
		UpdatedBy:   &inventoryType.UpdatedBy,
		UpdatedAt:   inventoryType.UpdatedAt.String(),
	}

	reponsePass := types.InventoryTypeResponse{
		Status: statusCode.Success("OK"),
		Data:   &inventoryTypeData,
	}

	return &reponsePass, nil
}
