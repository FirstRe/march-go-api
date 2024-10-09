package inventoryService

import (
	"errors"
	"log"
	"march-inventory/cmd/app/common"
	gormDb "march-inventory/cmd/app/common/gorm"
	"march-inventory/cmd/app/common/helper"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"

	"gorm.io/gorm"
)

const ClassName string = "InventoryService"

func UpsertInventoryType(input *types.UpsertInventoryTypeInput) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryType")
	logctx.Logger([]interface{}{input}, "input")

	findDup := model.InventoryType{}
	gormDb.Repos.Model(&model.InventoryType{}).Where("name = ?", input.Name).Find(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryResponse{
			Status: common.StatusResponse(400, "Duplicated"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryResponse{
			Status: common.StatusResponse(400, "Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventoryTypeData := model.InventoryType{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   "system",
		UpdatedBy:   "system",
	}

	if input.ID != nil {
		inventoryTypeData.ID = *input.ID
		inventoryTypeData.CreatedBy = findDup.CreatedBy
		inventoryTypeData.UpdatedBy = findDup.UpdatedBy
		inventoryTypeData.Deleted = findDup.Deleted
		log.Printf("input.ID: %+v", input.ID)
	}

	log.Printf("inventoryTypeData%+v", inventoryTypeData)

	if err := gormDb.InventoryType.Save(&inventoryTypeData).Error; err != nil {
		if errors.Is(err, gorm.ErrMissingWhereClause) {
			log.Printf("err ErrMissingWhereClause: %+v", err)
			if err := gormDb.InventoryType.Save(&inventoryTypeData).Where("id = ?", inventoryTypeData.ID).Error; err != nil {
				log.Printf("err Create: %+v", err)
			} else {
				reponsePass := types.MutationInventoryResponse{
					Status: common.StatusResponse(200, "OK"),
					Data: &types.ResponseID{
						ID: &inventoryTypeData.ID,
					},
				}
				return &reponsePass, nil
			}
		} else {
			log.Printf("err Create: %+v", err)

		}
		reponseError := types.MutationInventoryResponse{
			Status: common.StatusResponse(500, "CREATE ERROR"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	log.Printf("inventoryTypeData%+v", inventoryTypeData)

	reponsePass := types.MutationInventoryResponse{
		Status: common.StatusResponse(200, "OK"),
		Data: &types.ResponseID{
			ID: &inventoryTypeData.ID,
		},
	}
	return &reponsePass, nil
}

func DeleteInventoryType(id string) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryType")
	logctx.Logger([]interface{}{id}, "id")

	inventoryType := model.InventoryType{}
	if err := gormDb.Repos.Model(&model.InventoryType{}).Where("id = ?", id).First(&inventoryType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("InventoryType with id %s not found", id)
			reponseError := types.MutationInventoryResponse{
				Status: common.StatusResponse(404, "Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		log.Printf("Error fetching InventoryType: %+v", err)
		reponseError := types.MutationInventoryResponse{
			Status: common.StatusResponse(500, "Internal server error"),
			Data:   nil,
		}
		return &reponseError, err
	}

	if err := gormDb.Repos.Model(&inventoryType).Update("deleted", true).Error; err != nil {
		log.Printf("Error deleting InventoryType: %+v", err)
		reponseError := types.MutationInventoryResponse{
			Status: common.StatusResponse(500, "Error deleting inventory type"),
			Data:   nil,
		}
		return &reponseError, err
	}

	reponseSuccess := types.MutationInventoryResponse{
		Status: common.StatusResponse(200, "OK"),
		Data:   nil,
	}
	return &reponseSuccess, nil
}

func GetInventoryTypes(params *types.ParamsInventoryType) (*types.InventoryTypesResponse, error) {
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
			Name:        &inventoryType.Name,
			Description: inventoryType.Description,
			CreatedBy:   &inventoryType.CreatedBy,
			CreatedAt:   inventoryType.CreatedAt.String(),
			UpdatedBy:   &inventoryType.UpdatedBy,
			UpdatedAt:   inventoryType.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryTypesResponse{
		Status: common.StatusResponse(200, "OK"),
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
		code := 500
		message := "Internal Error"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code = 400
			message = "Not Found"
		}

		reponseError := types.InventoryTypeResponse{
			Status: common.StatusResponse(code, message),
			Data:   nil,
		}
		return &reponseError, nil
	}
	logctx.Logger([]interface{}{inventoryType}, "inventoryType")

	inventoryTypeData := types.InventoryType{
		ID:          &inventoryType.ID,
		Name:        &inventoryType.Name,
		Description: inventoryType.Description,
		CreatedBy:   &inventoryType.CreatedBy,
		CreatedAt:   inventoryType.CreatedAt.String(),
		UpdatedBy:   &inventoryType.UpdatedBy,
		UpdatedAt:   inventoryType.UpdatedAt.String(),
	}

	reponsePass := types.InventoryTypeResponse{
		Status: common.StatusResponse(200, "OK"),
		Data:   &inventoryTypeData,
	}

	return &reponsePass, nil
}
