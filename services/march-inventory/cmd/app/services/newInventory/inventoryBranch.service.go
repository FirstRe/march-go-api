package newInventory

import (
	"core/app/helper"
	"core/app/middlewares"
	"errors"
	"log"

	"march-inventory/cmd/app/common/statusCode"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/repositories"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (i inventoryServiceRedis) UpsertInventoryBranch(input *types.UpsertInventoryBranchInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBranch")
	logctx.Logger(input, "input")

	branchName := input.Name + "|" + userInfo.UserInfo.ShopsID

	findDup, _ := i.inventoryRepo.FindFirstInventoryBranch(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{
			Where: map[string]interface{}{
				"name":     branchName,
				"shops_Id": userInfo.UserInfo.ShopsID,
			}}}})

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if input.Name == "" {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest("name is required"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(findDup, "findDup")
	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryBranchResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	findDup = model.InventoryBranch{
		ID:          "",
		Name:        branchName,
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
	err := i.inventoryRepo.SaveInventoryBranch(findDup)

	if err != nil {
		logctx.Logger(err, "[error-api] Upsert")
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

func (i inventoryServiceRedis) DeleteInventoryBranch(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBranch")
	logctx.Logger(id, "id")

	inventoryBranch, err := i.inventoryRepo.FindFirstInventoryBranch(repositories.FindParams{WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}}})
	if err != nil {
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

	_, err = i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"inventory_branch_id": id}}},
	})

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryBranch not found")
		inventoryBranch.Deleted = true

		err := i.inventoryRepo.SaveInventoryBranch(inventoryBranch)
		if err != nil {
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

func (i inventoryServiceRedis) GetInventoryBranchs(params *types.ParamsInventoryBranch, userInfo middlewares.UserClaims) (*types.InventoryBranchsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBranchs")

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%|%"
	}
	log.Printf("searchParam: %+v", searchParam)

	inventoryBranchs, err := i.inventoryRepo.FindInventoryBranch(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: "name ILIKE ?", WhereArgs: []interface{}{searchParam}}, {Where: map[string]interface{}{"deleted": false}}},
		OrderBy:   "created_at asc",
	})

	if err != nil {
		logctx.Logger(err, "err GetInventoryBranchs Model Data")
	}
	logctx.Logger(inventoryBranchs, "inventoryBranchs")
	inventoryBranchsData := make([]*types.InventoryBranch, len(inventoryBranchs))

	for d, inventoryBranch := range inventoryBranchs {
		inventoryBranchsData[d] = &types.InventoryBranch{
			ID:          &inventoryBranch.ID,
			Name:        strings.Split(inventoryBranch.Name, "|")[0],
			Description: inventoryBranch.Description,
			CreatedBy:   &inventoryBranch.CreatedBy,
			CreatedAt:   inventoryBranch.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventoryBranch.UpdatedBy,
			UpdatedAt:   inventoryBranch.UpdatedAt.UTC().Format(time.DateTime),
		}
	}

	reponsePass := types.InventoryBranchsDataResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryBranchsData,
	}

	return &reponsePass, nil
}

func (i inventoryServiceRedis) GetInventoryBranch(id *string) (*types.InventoryBranch, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBranch")
	logctx.Logger(id, "id")

	inventoryBranch, err := i.inventoryRepo.FindFirstInventoryBranch(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{
			{Where: map[string]interface{}{"id": id}}},
	})

	if err != nil {
		logctx.Logger(err, "err GetInventoryBranch Model Data")

		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger(err, "not found")
		}
		return nil, nil
	}
	logctx.Logger(inventoryBranch, "InventoryBranch")

	inventoryBranchData := types.InventoryBranch{
		ID:          &inventoryBranch.ID,
		Name:        strings.Split(inventoryBranch.Name, "|")[0],
		Description: inventoryBranch.Description,
		CreatedBy:   &inventoryBranch.CreatedBy,
		CreatedAt:   inventoryBranch.CreatedAt.UTC().Format(time.DateTime),
		UpdatedBy:   &inventoryBranch.UpdatedBy,
		UpdatedAt:   inventoryBranch.UpdatedAt.UTC().Format(time.DateTime),
	}

	return &inventoryBranchData, nil
}
