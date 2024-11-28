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

func UpsertInventoryBranch(input *types.UpsertInventoryBranchInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBranch")
	logctx.Logger(input, "input")
	findDup := model.InventoryBranch{}
	typeName := input.Name + "|" + userInfo.UserInfo.ShopsID
	gormDb.Repos.Model(&model.InventoryBranch{}).Where("name = ? AND shops_Id = ?", typeName, userInfo.UserInfo.ShopsID).First(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(findDup, "findDup")
	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest("Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	findDup = model.InventoryBranch{
		ID:          "",
		Name:        typeName,
		Description: input.Description,
		ShopsID:     userInfo.UserInfo.ShopsID,
		CreatedBy:   userInfo.UserInfo.UserName,
		UpdatedBy:   userInfo.UserInfo.UserName,
	}

	onOkLocalT := "Upsert.success.create.branch"
	saveFailedLocalT := "Upsert.failed.create"

	if input.ID != nil {
		findDup.ID = *input.ID
		onOkLocalT = "Upsert.success.update.branch"
		saveFailedLocalT = "Upsert.failed.update"
		if findDup.ShopsID != userInfo.UserInfo.ShopsID {
			reponseError := types.MutationInventoryBranchResponse{
				Status: statusCode.Forbidden("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
	}

	logctx.Logger(findDup, "InventoryBranchData", true)

	if result := gormDb.Repos.Save(&findDup); result.Error != nil {
		logctx.Logger(result.Error, "[error-api] Upsert")
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.InternalError(translation.LocalizeMessage(saveFailedLocalT)),
			Data:   nil,
		}
		return &reponseError, nil
	} else {
		reponsePass := types.MutationInventoryBranchResponse{
			Status: statusCode.Success(translation.LocalizeMessage(onOkLocalT)),
			Data: &types.ResponseID{
				ID: &findDup.ID,
			},
		}
		return &reponsePass, nil
	}

}

func DeleteInventoryBranch(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBranch")
	logctx.Logger(id, "id")

	inventoryBranch := model.InventoryBranch{}
	if err := gormDb.Repos.Where("id = ?", id).First(&inventoryBranch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger(id, "[error-api] InventoryBranch with id not found")
			reponseError := types.MutationInventoryBranchResponse{
				Status: statusCode.BadRequest("Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		logctx.Logger(err.Error(), "[error-api] fetching InventoryBranch")

		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.InternalError(""),
			Data:   nil,
		}
		return &reponseError, err
	}
	logctx.Logger(inventoryBranch, "[log-api] inventoryBranch")
	if inventoryBranch.ShopsID != userInfo.UserInfo.ShopsID {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventory := model.Inventory{}
	result := gormDb.Repos.Where("inventory_branch_id = ?", id).First(&inventory)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryBranch not found")

		if err := gormDb.Repos.Model(&inventoryBranch).Update("deleted", true).Error; err != nil {
			logctx.Logger(id, "[error-api] deleting InventoryBranch")
			reponseError := types.MutationInventoryBranchResponse{
				Status: statusCode.InternalError("Error deleting inventory branch"),
				Data:   nil,
			}
			return &reponseError, err
		} else {
			reponseSuccess := types.MutationInventoryBranchResponse{
				Status: statusCode.Success(translation.LocalizeMessage("Success.branch")),
				Data:   nil,
			}
			return &reponseSuccess, nil
		}
	} else {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("onUse.branch")),
			Data:   nil,
		}
		return &reponseError, nil
	}

}

func GetInventoryBranchs(params *types.ParamsInventoryBranch, userInfo middlewares.UserClaims) (*types.InventoryBranchsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBranchs")
	logctx.Logger([]interface{}{}, "id")

	inventoryBranchs := []model.InventoryBranch{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if err := gormDb.Repos.Model(&inventoryBranchs).Where("name LIKE ?", searchParam).Not("deleted = ?", true).Order("created_at asc").Find(&inventoryBranchs).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryBranchs Model Data")
	}
	logctx.Logger([]interface{}{inventoryBranchs}, "inventoryBranchs")
	inventoryBranchsData := make([]*types.InventoryBranch, len(inventoryBranchs))

	for d, inventoryBranch := range inventoryBranchs {
		inventoryBranchsData[d] = &types.InventoryBranch{
			ID:          &inventoryBranch.ID,
			Name:        inventoryBranch.Name,
			Description: inventoryBranch.Description,
			CreatedBy:   &inventoryBranch.CreatedBy,
			CreatedAt:   inventoryBranch.CreatedAt.String(),
			UpdatedBy:   &inventoryBranch.UpdatedBy,
			UpdatedAt:   inventoryBranch.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryBranchsDataResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryBranchsData,
	}

	return &reponsePass, nil
}

func GetInventoryBranch(id *string) (*types.InventoryBranch, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBranch")
	logctx.Logger([]interface{}{id}, "id")
	inventoryBranch := model.InventoryBranch{}
	if err := gormDb.Repos.Model(&inventoryBranch).Where("id = ?", id).First(&inventoryBranch).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryType Model Data")

		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger([]interface{}{}, "not found")
		}
		return nil, nil
	}
	logctx.Logger([]interface{}{inventoryBranch}, "InventoryBranch")

	inventoryBranchData := types.InventoryBranch{
		ID:          &inventoryBranch.ID,
		Name:        inventoryBranch.Name,
		Description: inventoryBranch.Description,
		CreatedBy:   &inventoryBranch.CreatedBy,
		CreatedAt:   inventoryBranch.CreatedAt.String(),
		UpdatedBy:   &inventoryBranch.UpdatedBy,
		UpdatedAt:   inventoryBranch.UpdatedAt.String(),
	}

	return &inventoryBranchData, nil
}
