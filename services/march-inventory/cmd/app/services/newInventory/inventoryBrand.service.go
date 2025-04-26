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

func (i inventoryServiceRedis) UpsertInventoryBrand(input *types.UpsertInventoryBrandInput, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBrand")
	logctx.Logger(input, "input")
	typeName := input.Name + "|" + userInfo.UserInfo.ShopsID

	findDup, _ := i.inventoryRepo.FindFirstInventoryBrand(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{
			Where: map[string]interface{}{
				"name":     typeName,
				"shops_Id": userInfo.UserInfo.ShopsID,
			}}}})

	if input.Name == "" {
		reponseError := types.MutationInventoryBrandResponse{
			Status: statusCode.BadRequest("name is required"),
			Data:   nil,
		}
		return &reponseError, nil
	}

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
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
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

	err := i.inventoryRepo.SaveInventoryBrand(findDup)

	if err != nil {
		logctx.Logger(err, "[error-api] Upsert")
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

func (i inventoryServiceRedis) DeleteInventoryBrand(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBrand")
	logctx.Logger(id, "id")

	inventoryBrand, err := i.inventoryRepo.FindFirstInventoryBrand(repositories.FindParams{WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}}})
	if err != nil {
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

	_, err = i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"inventory_brand_id": id}}},
	})

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryBrand not found")
		inventoryBrand.Deleted = true

		err := i.inventoryRepo.SaveInventoryBrand(inventoryBrand)
		if err != nil {
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

func (i inventoryServiceRedis) GetInventoryBrands(params *types.ParamsInventoryBrand, userInfo middlewares.UserClaims) (*types.InventoryBrandsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBrands")
	logctx.Logger([]interface{}{}, "id")

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%|%"
	}
	log.Printf("searchParam: %+v", searchParam)

	inventoryBrands, err := i.inventoryRepo.FindInventoryBrand(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: "name ILIKE ?", WhereArgs: []interface{}{searchParam}}, {Where: map[string]interface{}{"deleted": false}}},
		OrderBy:   "created_at asc",
	})

	if err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryBrands Model Data")
	}
	logctx.Logger([]interface{}{inventoryBrands}, "inventoryBrands")
	inventoryBrandsData := make([]*types.InventoryBrand, len(inventoryBrands))

	for d, inventoryBrand := range inventoryBrands {
		inventoryBrandsData[d] = &types.InventoryBrand{
			ID:          &inventoryBrand.ID,
			Name:        strings.Split(inventoryBrand.Name, "|")[0],
			Description: inventoryBrand.Description,
			CreatedBy:   &inventoryBrand.CreatedBy,
			CreatedAt:   inventoryBrand.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventoryBrand.UpdatedBy,
			UpdatedAt:   inventoryBrand.UpdatedAt.UTC().Format(time.DateTime),
		}
	}

	reponsePass := types.InventoryBrandsDataResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryBrandsData,
	}

	return &reponsePass, nil
}

func (i inventoryServiceRedis) GetInventoryBrand(id *string) (*types.InventoryBrand, error) {
	//notuse
	logctx := helper.LogContext(ClassName, "GetInventoryBrand")
	logctx.Logger([]interface{}{id}, "id")

	inventoryBrand, err := i.inventoryRepo.FindFirstInventoryBrand(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{
			{Where: map[string]interface{}{"id": id}}},
	})

	if err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryBrand Model Data")

		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger([]interface{}{}, "not found")
		}
		return nil, nil
	}
	logctx.Logger([]interface{}{inventoryBrand}, "InventoryBrand")

	inventoryBrandData := types.InventoryBrand{
		ID:          &inventoryBrand.ID,
		Name:        strings.Split(inventoryBrand.Name, "|")[0],
		Description: inventoryBrand.Description,
		CreatedBy:   &inventoryBrand.CreatedBy,
		CreatedAt:   inventoryBrand.CreatedAt.UTC().Format(time.DateTime),
		UpdatedBy:   &inventoryBrand.UpdatedBy,
		UpdatedAt:   inventoryBrand.UpdatedAt.UTC().Format(time.DateTime),
	}

	return &inventoryBrandData, nil
}
