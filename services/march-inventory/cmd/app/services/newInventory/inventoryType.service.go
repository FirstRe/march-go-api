package newInventory

import (
	"context"
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

func (i inventoryServiceRedis) UpsertInventoryType(input *types.UpsertInventoryTypeInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryType")
	logctx.Logger(input, "input")
	typeName := input.Name + "|" + userInfo.UserInfo.ShopsID

	findDup, _ := i.inventoryRepo.FindFirstInventoryType(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: "name = ? AND shops_Id = ?", WhereArgs: []interface{}{typeName, userInfo.UserInfo.ShopsID}}},
	})

	if input.Name == "" {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("name is required"),
			Data:   nil,
		}
		return &reponseError, nil
	}

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
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
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

	err := i.inventoryRepo.SaveInventoryType(findDup)
	if err != nil {
		logctx.Logger(err.Error, "[error-api] Upsert")
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

func (i inventoryServiceRedis) DeleteInventoryType(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryType")
	logctx.Logger(id, "id")

	inventoryType, err := i.inventoryRepo.FindFirstInventoryType(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}},
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logctx.Logger(id, "[error-api] InventoryType with id not found")
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.BadRequest("Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		logctx.Logger(err.Error(), "[error-api] fetching InventoryType")

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
	_, err = i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"inventory_type_id": id}}},
	})

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		logctx.Logger(id, "Inventory with InventoryType not found")
		//update deleted
		inventoryType.Deleted = true
		logctx.Logger(inventoryType, "inventoryTypenot found")

		err = i.inventoryRepo.SaveInventoryType(inventoryType)

		if err != nil {
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

func (i inventoryServiceRedis) GetInventoryTypes(params *types.ParamsInventoryType, userInfo middlewares.UserClaims) (*types.InventoryTypesResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryTypes")
	_, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//grpc
	// shopIds := "984d0d87-7d74-45c5-9d94-6ebcb74a98de"
	// r, err := grpcAuth.GetPermission(shopIds, "token")
	// if err != nil {
	// 	log.Fatalf("could not greet: %v", err)
	// }
	// log.Printf("Greeting22: %s", r.GetShop())

	// inventoryTypes := []model.InventoryType{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%|%"
	}
	log.Printf("searchParam: %+v", searchParam)

	inventoryTypes, err := i.inventoryRepo.FindInventoryType(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: "name ILIKE ?", WhereArgs: []interface{}{searchParam}}, {Where: map[string]interface{}{"deleted": false}}},
		OrderBy:   "created_at asc",
	})

	if err != nil {
		logctx.Logger(err, "err GetInventoryTypes Model Data")
	}
	logctx.Logger(inventoryTypes, "inventoryTypes")
	inventoryTypesData := make([]*types.InventoryType, len(inventoryTypes))

	for d, inventoryType := range inventoryTypes {
		inventoryTypesData[d] = &types.InventoryType{
			ID:          &inventoryType.ID,
			Name:        strings.Split(inventoryType.Name, "|")[0],
			Description: inventoryType.Description,
			CreatedBy:   &inventoryType.CreatedBy,
			CreatedAt:   inventoryType.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventoryType.UpdatedBy,
			UpdatedAt:   inventoryType.UpdatedAt.UTC().Format(time.DateTime),
		}
	}

	reponsePass := types.InventoryTypesResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryTypesData,
	}

	return &reponsePass, nil
}

func (i inventoryServiceRedis) GetInventoryType(id *string) (*types.InventoryTypeResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryType")
	logctx.Logger(id, "id")
	inventoryType, err := i.inventoryRepo.FindFirstInventoryType(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{
			{Where: map[string]interface{}{"id": id}}},
	})

	if err != nil {
		logctx.Logger(err, "err GetInventoryType Model Data")
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
	logctx.Logger(inventoryType, "inventoryType")

	inventoryTypeData := types.InventoryType{
		ID:          &inventoryType.ID,
		Name:        strings.Split(inventoryType.Name, "|")[0],
		Description: inventoryType.Description,
		CreatedBy:   &inventoryType.CreatedBy,
		CreatedAt:   inventoryType.CreatedAt.UTC().Format(time.DateTime),
		UpdatedBy:   &inventoryType.UpdatedBy,
		UpdatedAt:   inventoryType.UpdatedAt.UTC().Format(time.DateTime),
	}

	reponsePass := types.InventoryTypeResponse{
		Status: statusCode.Success("OK"),
		Data:   &inventoryTypeData,
	}

	return &reponsePass, nil
}
