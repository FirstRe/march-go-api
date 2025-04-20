package repositories

import (
	. "core/app/helper"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inventoryRepositoryDB struct {
	gormDb *gorm.DB
}

func NewProductRepository(db *gorm.DB) InventoryRepository {
	return inventoryRepositoryDB{db}
}

func (r inventoryRepositoryDB) GetInventories(searchParam string, isSerialNumber bool, params *types.ParamsInventory, shopId string) (inventory []model.Inventory, totalRow int, err error) {
	pageNo := DefaultTo(params.PageNo, 1)
	limit := DefaultTo(params.Limit, 30)
	offset := pageNo*limit - limit

	query := r.gormDb.Model(&inventory).Where("deleted = ?", false)

	if searchParam != "" && !isSerialNumber {
		query = query.Where("name ILIKE ?", searchParam)
	} else if isSerialNumber {
		query = query.Where("serial_number ILIKE ?", searchParam)
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

	query = query.Where("shops_id = ?", shopId)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	totalRow = int(count)

	if err := query.Preload(clause.Associations).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&inventory).Error; err != nil {
		return nil, 0, err
	}

	return inventory, totalRow, nil
}

func (r inventoryRepositoryDB) FindFirstInventory(where map[string]interface{}) (inventory model.Inventory, err error) {
	if err := r.gormDb.Where(where).First(&inventory).Error; err != nil {
		return model.Inventory{}, err
	}
	return inventory, nil
}

func (r inventoryRepositoryDB) UpdateInventory(id string, updatedData map[string]interface{}) error {
	if err := r.gormDb.Model(&model.Inventory{}).
		Where("id = ?", id).
		Updates(updatedData).Error; err != nil {
		return err
	}
	return nil
}

func (r inventoryRepositoryDB) SaveInventory(inventoryData model.Inventory) error {
	if err := r.gormDb.Omit("CreatedAt").Save(&inventoryData).Error; err != nil {
		return err
	}
	return nil
}
