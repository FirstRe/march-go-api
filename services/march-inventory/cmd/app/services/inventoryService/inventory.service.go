package inventoryService

import (
	. "core/app/helper"
	"core/app/middlewares"
	"errors"
	"log"
	"march-inventory/cmd/app/dto"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	"march-inventory/cmd/app/statusCode"
	gormDb "march-inventory/cmd/app/statusCode/gorm"
	"math"
	"strings"

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

func GetInventories(params *types.ParamsInventory, userInfo middlewares.UserClaims) (*types.InventoriesResponse, error) {
	logctx := LogContext(ClassName, "GetInventories")
	logctx.Logger(params, "params")

	pageNo := DefaultTo(params.PageNo, 1)
	limit := DefaultTo(params.Limit, 30)
	offset := pageNo*limit - limit

	logctx.Logger(offset, "offset")
	logctx.Logger(pageNo, "pageNo")
	logctx.Logger(limit, "limit")

	inventories := []model.Inventory{}

	searchParam := ""
	isSerialNumber := false
	if params != nil && params.Search != nil {
		if strings.HasPrefix(*params.Search, "#") {
			isSerialNumber = true
		}
		searchParam = "%" + *params.Search + "%"
	}
	log.Printf("searchParam: %+v", searchParam)

	query := gormDb.Repos.Model(&[]model.Inventory{}).Where("deleted = ?", false)

	if searchParam != "" && !isSerialNumber {
		query = query.Where("name LIKE ?", searchParam)
	} else if isSerialNumber {
		query = query.Where("serial_number LIKE ?", searchParam)
	}

	if params.Favorite != nil && *params.Favorite == types.FavoriteStatusLike {
		query = query.Where("favorite = ?", true)
	}

	if len(params.Type) > 0 {
		query = query.Where("inventory_type_id IN ?", params.Type).Preload("InventoryType")
	}

	if len(params.Brand) > 0 {
		query = query.Where("inventory_brand_id IN ?", params.Brand).Preload("InventoryBrand")
	}

	if len(params.Branch) > 0 {
		query = query.Where("inventory_branch_id IN ?", params.Branch).Preload("InventoryBranch")
	}

	if userInfo.UserInfo.ShopsID == "" || userInfo.UserInfo.UserName == "" {
		reponseError := types.InventoriesResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	query = query.Where("shops_id = ?", userInfo.UserInfo.ShopsID)

	result := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&inventories)

	if result.Error != nil {
		logctx.Logger([]interface{}{result.Error}, "err GetInventories Model Data")
		reponseError := types.InventoriesResponse{
			Status: statusCode.InternalError(""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(inventories, "[log-api] inventories")

	inventoriesData := make([]*types.Inventory, len(inventories))

	for d, inventory := range inventories {
		inventoryBrand := types.InventoryBrand{
			ID:          &inventory.InventoryBrand.ID,
			Name:        inventory.InventoryBrand.Name,
			Description: inventory.InventoryBrand.Description,
			CreatedBy:   &inventory.InventoryBrand.CreatedBy,
			CreatedAt:   inventory.InventoryBrand.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryBrand.UpdatedBy,
			UpdatedAt:   inventory.InventoryBrand.UpdatedAt.String(),
		}

		inventoryBranch := types.InventoryBranch{
			ID:          &inventory.InventoryBranch.ID,
			Name:        inventory.InventoryBranch.Name,
			Description: inventory.InventoryBranch.Description,
			CreatedBy:   &inventory.InventoryBranch.CreatedBy,
			CreatedAt:   inventory.InventoryBranch.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryBranch.UpdatedBy,
			UpdatedAt:   inventory.InventoryBranch.UpdatedAt.String(),
		}

		inventoryType := types.InventoryType{
			ID:          &inventory.InventoryType.ID,
			Name:        inventory.InventoryType.Name,
			Description: inventory.InventoryType.Description,
			CreatedBy:   &inventory.InventoryType.CreatedBy,
			CreatedAt:   inventory.InventoryType.CreatedAt.String(),
			UpdatedBy:   &inventory.InventoryType.UpdatedBy,
			UpdatedAt:   inventory.InventoryType.UpdatedAt.String(),
		}

		inventoriesData[d] = &types.Inventory{
			ID:              &inventory.ID,
			Name:            inventory.Name,
			Description:     inventory.Description,
			CreatedBy:       &inventory.CreatedBy,
			CreatedAt:       inventory.CreatedAt.String(),
			UpdatedBy:       &inventory.UpdatedBy,
			UpdatedAt:       inventory.UpdatedAt.String(),
			Amount:          inventory.Amount,
			Sold:            &inventory.Sold,
			Sku:             inventory.SKU,
			SerialNumber:    inventory.SerialNumber,
			Size:            inventory.Size,
			PriceMember:     inventory.PriceMember,
			Price:           inventory.Price,
			ReorderLevel:    inventory.ReorderLevel,
			ExpiryDate:      inventory.ExpiryDate.String(),
			InventoryBrand:  &inventoryBrand,
			InventoryBranch: &inventoryBranch,
			InventoryType:   &inventoryType,
			Favorite:        &inventory.Favorite,
		}
	}
	var count int64
	if err := query.Find(&inventories).Count(&count).Error; err != nil {
		count = 0
	}
	totalRow := int(count)
	totalPage := int(math.Ceil(float64(totalRow) / float64(limit)))

	reponseData := types.ResponseInventories{
		Inventories: inventoriesData,
		PageLimit:   params.Limit,
		TotalRow:    &totalRow,
		TotalPage:   &totalPage,
		PageNo:      params.PageNo,
	}

	reponsePass := types.InventoriesResponse{
		Status: statusCode.Success("OK"),
		Data:   &reponseData,
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
