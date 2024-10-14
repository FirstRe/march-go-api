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

func UpsertInventoryBranch(input *types.UpsertInventoryBranchInput) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventoryBranch")
	logctx.Logger([]interface{}{input}, "input")

	findDup := model.InventoryBranch{}
	gormDb.Repos.Model(&model.InventoryBranch{}).Where("name = ?", input.Name).Find(&findDup)

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryBranchResponse{
			Status: common.StatusResponse(400, "Duplicated"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryBranchResponse{
			Status: common.StatusResponse(400, "Bad Request"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	inventoryBranchData := model.InventoryBranch{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   "system",
		UpdatedBy:   "system",
	}

	if input.ID != nil {
		inventoryBranchData.ID = *input.ID
		inventoryBranchData.CreatedBy = findDup.CreatedBy
		inventoryBranchData.UpdatedBy = findDup.UpdatedBy
		inventoryBranchData.Deleted = findDup.Deleted
		log.Printf("input.ID: %+v", input.ID)
	}

	log.Printf("inventoryBranchData%+v", inventoryBranchData)

	if err := gormDb.Repos.Model(&model.InventoryBranch{}).Save(&inventoryBranchData).Error; err != nil {
		if errors.Is(err, gorm.ErrMissingWhereClause) {
			log.Printf("err ErrMissingWhereClause: %+v", err)
			if err := gormDb.Repos.Model(&model.InventoryBranch{}).Save(&inventoryBranchData).Where("id = ?", inventoryBranchData.ID).Error; err != nil {
				log.Printf("err Create: %+v", err)
			} else {
				reponsePass := types.MutationInventoryBranchResponse{
					Status: common.StatusResponse(1000, "OK"),
					Data: &types.ResponseID{
						ID: &inventoryBranchData.ID,
					},
				}
				return &reponsePass, nil
			}
		} else {
			log.Printf("err Create: %+v", err)

		}
		reponseError := types.MutationInventoryBranchResponse{
			Status: common.StatusResponse(500, "CREATE ERROR"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	log.Printf("inventoryBranchData: %+v", inventoryBranchData)

	reponsePass := types.MutationInventoryBranchResponse{
		Status: common.StatusResponse(1000, "OK"),
		Data: &types.ResponseID{
			ID: &inventoryBranchData.ID,
		},
	}
	return &reponsePass, nil
}

func DeleteInventoryBranch(id string) (*types.MutationInventoryBranchResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventoryBranch")
	logctx.Logger([]interface{}{id}, "id")

	inventoryBranch := model.InventoryBranch{}
	if err := gormDb.Repos.Model(&model.InventoryBranch{}).Where("id = ?", id).First(&inventoryBranch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("InventoryBranch with id %s not found", id)
			reponseError := types.MutationInventoryBranchResponse{
				Status: common.StatusResponse(404, "Not found"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		log.Printf("Error fetching InventoryBranch: %+v", err)
		reponseError := types.MutationInventoryBranchResponse{
			Status: common.StatusResponse(500, "Internal server error"),
			Data:   nil,
		}
		return &reponseError, err
	}

	if err := gormDb.Repos.Model(&inventoryBranch).Update("deleted", true).Error; err != nil {
		log.Printf("Error deleting InventoryBranch: %+v", err)
		reponseError := types.MutationInventoryBranchResponse{
			Status: common.StatusResponse(500, "Error deleting inventory type"),
			Data:   nil,
		}
		return &reponseError, err
	}

	reponseSuccess := types.MutationInventoryBranchResponse{
		Status: common.StatusResponse(1000, "OK"),
		Data:   nil,
	}
	return &reponseSuccess, nil
}

func GetInventoryBranchs(params *types.ParamsInventoryBranch) (*types.InventoryBranchsDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventoryBranchs")
	logctx.Logger([]interface{}{}, "id")
	inventoryBranchs := []model.InventoryBranch{}

	searchParam := ""
	if params != nil && params.Search != nil {
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if err := gormDb.Repos.Model(&inventoryBranchs).Where("name LIKE ?", searchParam).Not("deleted = ?", true).Order("created_at asc").Find(&inventoryBranchs).Error; err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventoryTypes Model Data")
	}
	logctx.Logger([]interface{}{inventoryBranchs}, "InventoryBranch")
	inventoryBranchsData := make([]*types.InventoryBranch, len(inventoryBranchs))

	for d, InventoryBranch := range inventoryBranchs {
		inventoryBranchsData[d] = &types.InventoryBranch{
			ID:          &InventoryBranch.ID,
			Name:        &InventoryBranch.Name,
			Description: InventoryBranch.Description,
			CreatedBy:   &InventoryBranch.CreatedBy,
			CreatedAt:   InventoryBranch.CreatedAt.String(),
			UpdatedBy:   &InventoryBranch.UpdatedBy,
			UpdatedAt:   InventoryBranch.UpdatedAt.String(),
		}
	}

	reponsePass := types.InventoryBranchsDataResponse{
		Status: common.StatusResponse(1000, "OK"),
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
		Name:        &inventoryBranch.Name,
		Description: inventoryBranch.Description,
		CreatedBy:   &inventoryBranch.CreatedBy,
		CreatedAt:   inventoryBranch.CreatedAt.String(),
		UpdatedBy:   &inventoryBranch.UpdatedBy,
		UpdatedAt:   inventoryBranch.UpdatedAt.String(),
	}

	return &inventoryBranchData, nil
}
