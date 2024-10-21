package inventoryService

import (
	. "core/app/helper"
	"errors"
	"log"
	"march-inventory/cmd/app/dto"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	"march-inventory/cmd/app/statusCode"
	gormDb "march-inventory/cmd/app/statusCode/gorm"

	"gorm.io/gorm"
)

func UpsertInventory(input types.UpsertInventoryInput) (*types.MutationInventoryResponse, error) {
	logctx := LogContext(ClassName, "UpsertInventory")
	logctx.Logger([]interface{}{input}, "input")

	findDup := model.Inventory{}
	gormDb.Repos.Model(&model.Inventory{}).Where("name = ?", input.Name).Find(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.Duplicated("Duplicated"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}
	log.Printf("inventoryData: %+v", input)

	inventoryData := dto.MapInputToInventory(input)

	log.Printf("inventoryData: %+v", inventoryData)
	if input.ID != nil {
		inventoryData.ID = *input.ID
		log.Printf("input.ID: %+v", input.ID)
	}

	log.Printf("inventoryData: %+v", inventoryData)

	if err := gormDb.Repos.Save(&inventoryData).Error; err != nil {
		if errors.Is(err, gorm.ErrMissingWhereClause) {
			log.Printf("err ErrMissingWhereClause: %+v", err)
			if err := gormDb.Repos.Save(&inventoryData).Where("id = ?", inventoryData.ID).Error; err != nil {
				log.Printf("err Create: %+v", err)
			} else {
				reponsePass := types.MutationInventoryResponse{
					Status: statusCode.Success("OK"),
					Data: &types.ResponseID{
						ID: &inventoryData.ID,
					},
				}
				return &reponsePass, nil
			}
		} else if errors.Is(err, gorm.ErrForeignKeyViolated) {
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.BadRequest("Bad Request"),
				Data:   nil,
			}
			return &reponseError, nil
		} else {
			log.Printf("err Create: %+v", err)
		}
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("CREATE ERROR"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	log.Printf("inventoryData: %+v", inventoryData)

	reponsePass := types.MutationInventoryResponse{
		Status: statusCode.Success("OK"),
		Data: &types.ResponseID{
			ID: &inventoryData.ID,
		},
	}
	return &reponsePass, nil
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

func GetInventoryTypess(params *types.ParamsInventoryType) (*types.InventoryTypesResponse, error) {
	logctx := LogContext(ClassName, "GetInventoryTypes")
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
		Status: statusCode.Success("OK"),
		Data:   inventoryTypesData,
	}

	return &reponsePass, nil
}

func GetInventoryTypesss(id *string) (*types.InventoryTypeResponse, error) {
	logctx := LogContext(ClassName, "GetInventoryType")
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
		Name:        &inventoryType.Name,
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
