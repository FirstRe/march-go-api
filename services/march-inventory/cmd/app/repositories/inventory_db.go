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

func (r inventoryRepositoryDB) mapParams(params FindParams) *gorm.DB {
	query := r.gormDb

	for _, whereArgs := range params.WhereArgs {
		if whereArgs.Where != nil {
			switch where := whereArgs.Where.(type) {
			case string:
				query = query.Where(where, whereArgs.WhereArgs...)
			case map[string]interface{}:
				query = query.Where(where)
			default:
				query = query.Where(where)
			}
		}

	}

	for _, preload := range params.Preload {
		query = query.Preload(preload)
	}

	if params.SelectField != nil {
		query = query.Select(params.SelectField)
	}

	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	}

	if params.Limit != nil {
		query = query.Limit(*params.Limit)
	}

	if params.Offset != nil {
		query = query.Offset(*params.Offset)
	}

	return query
}

func (r inventoryRepositoryDB) FindInventory(params FindParams) (inventories []model.Inventory, err error) {
	query := r.mapParams(params)
	if err := query.Find(&inventories).Error; err != nil {
		return nil, err
	}
	return inventories, nil
}

func (r inventoryRepositoryDB) FindFirstInventory(params FindParams) (inventory model.Inventory, err error) {
	query := r.mapParams(params)
	if err := query.First(&inventory).Error; err != nil {
		return model.Inventory{}, err
	}
	return inventory, nil
}

func (r inventoryRepositoryDB) UpdateInventory(id string, updatedData map[string]interface{}) error {
	return r.gormDb.Model(&model.Inventory{}).
		Where("id = ?", id).
		Updates(updatedData).Error
}

func (r inventoryRepositoryDB) SaveInventory(inventoryData model.Inventory) error {
	return r.gormDb.Omit("CreatedAt").Save(&inventoryData).Error
}

func (r inventoryRepositoryDB) DeleteSubInventory(checkIn interface{}, id string) error {
	return r.gormDb.Where("id = ?", id).Delete(&checkIn).Error
}

func (r inventoryRepositoryDB) RecoverySubInventory(checkIn interface{}, id string, updatedData map[string]interface{}, updatedBy string) error {
	return r.gormDb.Model(&checkIn).Where("id = ?", id).Updates(map[string]interface{}{"deleted": false, "updated_by": updatedBy}).Error
}
