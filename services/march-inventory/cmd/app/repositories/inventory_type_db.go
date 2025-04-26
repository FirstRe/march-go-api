package repositories

import "march-inventory/cmd/app/graph/model"

func (r inventoryRepositoryDB) FindInventoryType(params FindParams) (inventoryTypes []model.InventoryType, err error) {
	query := r.mapParams(params)
	if err := query.Find(&inventoryTypes).Error; err != nil {
		return nil, err
	}
	return inventoryTypes, nil
}

func (r inventoryRepositoryDB) FindFirstInventoryType(params FindParams) (inventoryType model.InventoryType, err error) {
	query := r.mapParams(params)
	if err := query.First(&inventoryType).Error; err != nil {
		return model.InventoryType{}, err
	}
	return inventoryType, nil
}

func (r inventoryRepositoryDB) SaveInventoryType(inventoryTypeData model.InventoryType) error {
	return r.gormDb.Omit("CreatedAt").Save(&inventoryTypeData).Error
}
