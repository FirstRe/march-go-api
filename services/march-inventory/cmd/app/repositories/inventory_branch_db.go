package repositories

import "march-inventory/cmd/app/graph/model"

func (r inventoryRepositoryDB) FindInventoryBranch(params FindParams) (inventoryBranchs []model.InventoryBranch, err error) {
	query := r.mapParams(params)
	if err := query.Find(&inventoryBranchs).Error; err != nil {
		return nil, err
	}
	return inventoryBranchs, nil
}

func (r inventoryRepositoryDB) FindFirstInventoryBranch(params FindParams) (inventoryBranch model.InventoryBranch, err error) {
	query := r.mapParams(params)
	if err := query.First(&inventoryBranch).Error; err != nil {
		return model.InventoryBranch{}, err
	}
	return inventoryBranch, nil
}

func (r inventoryRepositoryDB) SaveInventoryBranch(inventoryBranchData model.InventoryBranch) error {
	return r.gormDb.Omit("CreatedAt").Save(&inventoryBranchData).Error
}
