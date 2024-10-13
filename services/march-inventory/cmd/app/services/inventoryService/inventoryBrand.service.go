package inventoryService

import (
	"core/app/helper"
	"errors"
	"log"
	"march-inventory/cmd/app/common"
	gormDb "march-inventory/cmd/app/common/gorm"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"

	"gorm.io/gorm"
)

func UpsertInventoryBrand(input *types.UpsertInventoryBrandInput) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBrand")
	logctx.Logger([]interface{}{input}, "input")

	findDup := model.InventoryBrand{}
	gormDb.Repos.Model(&model.InventoryBrand{}).Where("name = ?", input.Name).Find(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryBrandResponse{
			Status: common.StatusResponse(400, "Duplicated"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryBrandResponse{
			Status: common.StatusResponse(400, "Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventoryBrandData := model.InventoryBrand{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   "system",
		UpdatedBy:   "system",
	}

	if input.ID != nil {
		inventoryBrandData.ID = *input.ID
		inventoryBrandData.CreatedBy = findDup.CreatedBy
		inventoryBrandData.UpdatedBy = findDup.UpdatedBy
		inventoryBrandData.Deleted = findDup.Deleted
		log.Printf("input.ID: %+v", input.ID)
	}

	log.Printf("inventoryBrandData%+v", inventoryBrandData)

	if err := gormDb.Repos.Model(&model.InventoryBrand{}).Save(&inventoryBrandData).Error; err != nil {
		if errors.Is(err, gorm.ErrMissingWhereClause) {
			log.Printf("err ErrMissingWhereClause: %+v", err)
			if err := gormDb.Repos.Model(&model.InventoryBrand{}).Save(&inventoryBrandData).Where("id = ?", inventoryBrandData.ID).Error; err != nil {
				log.Printf("err Create: %+v", err)
			} else {
				reponsePass := types.MutationInventoryBrandResponse{
					Status: common.StatusResponse(200, "OK"),
					Data: &types.ResponseID{
						ID: &inventoryBrandData.ID,
					},
				}
				return &reponsePass, nil
			}
		} else {
			log.Printf("err Create: %+v", err)

		}
		reponseError := types.MutationInventoryBrandResponse{
			Status: common.StatusResponse(500, "CREATE ERROR"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	log.Printf("inventoryBrandData: %+v", inventoryBrandData)

	reponsePass := types.MutationInventoryBrandResponse{
		Status: common.StatusResponse(200, "OK"),
		Data: &types.ResponseID{
			ID: &inventoryBrandData.ID,
		},
	}
	return &reponsePass, nil
}

func DeleteInventoryBrand(id string) (*types.MutationInventoryBrandResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBrand")
	logctx.Logger([]interface{}{id}, "id")

	inventoryBrand := model.InventoryBrand{}
	if err := gormDb.Repos.Model(&model.InventoryBrand{}).Where("id = ?", id).First(&inventoryBrand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("InventoryBrand with id %s not found", id)
			reponseError := types.MutationInventoryBrandResponse{
				Status: common.StatusResponse(404, "Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		log.Printf("Error fetching InventoryBrand: %+v", err)
		reponseError := types.MutationInventoryBrandResponse{
			Status: common.StatusResponse(500, "Internal server error"),
			Data:   nil,
		}
		return &reponseError, err
	}

	if err := gormDb.Repos.Model(&inventoryBrand).Update("deleted", true).Error; err != nil {
		log.Printf("Error deleting InventoryBrand: %+v", err)
		reponseError := types.MutationInventoryBrandResponse{
			Status: common.StatusResponse(500, "Error deleting inventory type"),
			Data:   nil,
		}
		return &reponseError, err
	}

	reponseSuccess := types.MutationInventoryBrandResponse{
		Status: common.StatusResponse(200, "OK"),
		Data:   nil,
	}
	return &reponseSuccess, nil
}

func GetInventoryBrands(params *types.ParamsInventoryBrand) (*types.InventoryBrandsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBrands")
	logctx.Logger([]interface{}{}, "id")
	inventoryBrands := []model.InventoryBrand{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if err := gormDb.Repos.Model(&inventoryBrands).Where("name LIKE ?", searchParam).Not("deleted = ?", true).Order("created_at asc").Find(&inventoryBrands).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryTypes Model Data")
	}
	logctx.Logger([]interface{}{inventoryBrands}, "InventoryBrand")
	inventoryBrandsData := make([]*types.InventoryBrand, len(inventoryBrands))

	for d, InventoryBrand := range inventoryBrands {
		inventoryBrandsData[d] = &types.InventoryBrand{
			ID:          &InventoryBrand.ID,
			Name:        &InventoryBrand.Name,
			Description: InventoryBrand.Description,
			CreatedBy:   &InventoryBrand.CreatedBy,
			CreatedAt:   InventoryBrand.CreatedAt.String(),
			UpdatedBy:   &InventoryBrand.UpdatedBy,
			UpdatedAt:   InventoryBrand.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryBrandsDataResponse{
		Status: common.StatusResponse(200, "OK"),
		Data:   inventoryBrandsData,
	}

	return &reponsePass, nil
}

func GetInventoryBrand(id *string) (*types.InventoryBrand, error) {
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
		Name:        &inventoryBrand.Name,
		Description: inventoryBrand.Description,
		CreatedBy:   &inventoryBrand.CreatedBy,
		CreatedAt:   inventoryBrand.CreatedAt.String(),
		UpdatedBy:   &inventoryBrand.UpdatedBy,
		UpdatedAt:   inventoryBrand.UpdatedAt.String(),
	}

	return &inventoryBrandData, nil
}
