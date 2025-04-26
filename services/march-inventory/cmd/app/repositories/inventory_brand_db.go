package repositories

import "march-inventory/cmd/app/graph/model"

func (r inventoryRepositoryDB) FindInventoryBrand(params FindParams) (inventoryBrands []model.InventoryBrand, err error) {
	query := r.mapParams(params)
	if err := query.Find(&inventoryBrands).Error; err != nil {
		return nil, err
	}
	return inventoryBrands, nil
}

func (r inventoryRepositoryDB) FindFirstInventoryBrand(params FindParams) (inventoryBrand model.InventoryBrand, err error) {
	query := r.mapParams(params)
	if err := query.First(&inventoryBrand).Error; err != nil {
		return model.InventoryBrand{}, err
	}
	return inventoryBrand, nil
}
