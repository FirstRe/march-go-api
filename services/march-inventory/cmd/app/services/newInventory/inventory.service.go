package newInventory

import (
	"context"
	"core/app/helper"
	. "core/app/helper"
	"core/app/middlewares"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"march-inventory/cmd/app/common/statusCode"
	"march-inventory/cmd/app/dto"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/graph/types"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/repositories"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const ClassName string = "InventoryService2"

type inventoryServiceRedis struct {
	redisClient   *redis.Client
	inventoryRepo repositories.InventoryRepository
}

func NewInventoryServiceRedis(redisClient *redis.Client, inventoryRepo repositories.InventoryRepository) InventoryService {
	return inventoryServiceRedis{
		redisClient:   redisClient,
		inventoryRepo: inventoryRepo,
	}
}

func (i inventoryServiceRedis) DeleteInventoryCache(key string) error {
	ctx := context.Background()
	var cursor uint64
	var batchSize int64 = 100

	for {
		keys, newCursor, err := i.redisClient.Scan(ctx, cursor, key+":*", batchSize).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := i.redisClient.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func generateInventoryCacheKey(shopsId string, pageNo, limit int, searchParam string, isSerialNumber bool) string {
	hasher := sha256.New()
	hasher.Write([]byte(searchParam))
	searchHash := hex.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("inventory:shopsId:%s:page:%d:limit:%d:search:%s:serial:%t", shopsId, pageNo, limit, searchHash, isSerialNumber)
}

func (i inventoryServiceRedis) GetInventories(params *types.ParamsInventory, userInfo middlewares.UserClaims) (*types.InventoriesResponse, error) {
	logctx := LogContext(ClassName, "GetInventories")
	logctx.Logger(params, "params")

	pageNo := DefaultTo(params.PageNo, 1)
	limit := DefaultTo(params.Limit, 30)
	offset := pageNo*limit - limit

	logctx.Logger(offset, "offset")
	logctx.Logger(pageNo, "pageNo")
	logctx.Logger(limit, "limit")

	searchParam := ""
	isSerialNumber := false

	if params != nil && params.Search != nil {
		if strings.HasPrefix(*params.Search, "#") {
			isSerialNumber = true
		}
		searchParam = "%" + *params.Search + "%|%"
	}
	log.Printf("searchParam: %+v", searchParam)

	if userInfo.UserInfo.ShopsID == "" || userInfo.UserInfo.UserName == "" {
		reponseError := types.InventoriesResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	key := generateInventoryCacheKey(userInfo.UserInfo.ShopsID, pageNo, limit, searchParam, isSerialNumber)

	if productsJson, err := i.redisClient.Get(context.Background(), key).Result(); err == nil {
		reponseData := types.ResponseInventories{}
		if err := json.Unmarshal([]byte(productsJson), &reponseData); err == nil {
			fmt.Println("getFrom redis")
			reponsePass := types.InventoriesResponse{
				Status: statusCode.Success("OK"),
				Data:   &reponseData,
			}
			return &reponsePass, nil
		}
	}

	logctx.Logger(key, "keynaja")
	inventories, totalRow, err := i.inventoryRepo.GetInventories(searchParam, isSerialNumber, params, userInfo.UserInfo.ShopsID)

	if err != nil {
		logctx.Logger([]interface{}{err}, "err GetInventories Model Data")
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
			Name:        strings.Split(inventory.InventoryBrand.Name, "|")[0],
			Description: inventory.InventoryBrand.Description,
			CreatedBy:   &inventory.InventoryBrand.CreatedBy,
			CreatedAt:   inventory.InventoryBrand.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryBrand.UpdatedBy,
			UpdatedAt:   inventory.InventoryBrand.UpdatedAt.UTC().Format(time.DateTime),
		}

		inventoryBranch := types.InventoryBranch{
			ID:          &inventory.InventoryBranch.ID,
			Name:        strings.Split(inventory.InventoryBranch.Name, "|")[0],
			Description: inventory.InventoryBranch.Description,
			CreatedBy:   &inventory.InventoryBranch.CreatedBy,
			CreatedAt:   inventory.InventoryBranch.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryBranch.UpdatedBy,
			UpdatedAt:   inventory.InventoryBranch.UpdatedAt.UTC().Format(time.DateTime),
		}

		inventoryType := types.InventoryType{
			ID:          &inventory.InventoryType.ID,
			Name:        strings.Split(inventory.InventoryType.Name, "|")[0],
			Description: inventory.InventoryType.Description,
			CreatedBy:   &inventory.InventoryType.CreatedBy,
			CreatedAt:   inventory.InventoryType.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryType.UpdatedBy,
			UpdatedAt:   inventory.InventoryType.UpdatedAt.UTC().Format(time.DateTime),
		}

		expiryDateStr := ""
		if inventory.ExpiryDate != nil {
			expiryDateStr = inventory.ExpiryDate.UTC().Format(time.DateTime)
		}

		inventoriesData[d] = &types.Inventory{
			ID:              &inventory.ID,
			Name:            strings.Split(inventory.Name, "|")[0],
			Description:     inventory.Description,
			CreatedBy:       &inventory.CreatedBy,
			CreatedAt:       inventory.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:       &inventory.UpdatedBy,
			UpdatedAt:       inventory.UpdatedAt.UTC().Format(time.DateTime),
			Amount:          inventory.Amount,
			Sold:            &inventory.Sold,
			Sku:             inventory.SKU,
			SerialNumber:    inventory.SerialNumber,
			Size:            inventory.Size,
			PriceMember:     inventory.PriceMember,
			Price:           inventory.Price,
			ReorderLevel:    inventory.ReorderLevel,
			ExpiryDate:      &expiryDateStr,
			InventoryBrand:  &inventoryBrand,
			InventoryBranch: &inventoryBranch,
			InventoryType:   &inventoryType,
			Favorite:        &inventory.Favorite,
		}
	}
	logctx.Logger(inventories, "[log-api] inventories1")

	totalPage := int(math.Ceil(float64(totalRow) / float64(limit)))

	logctx.Logger(totalRow, "[log-api] totalRow")
	logctx.Logger(totalPage, "[log-api] totalPage")

	reponseData := types.ResponseInventories{
		Inventories: inventoriesData,
		PageLimit:   params.Limit,
		TotalRow:    &totalRow,
		TotalPage:   &totalPage,
		PageNo:      params.PageNo,
	}

	if data, err := json.Marshal(reponseData); err == nil {
		i.redisClient.Set(context.Background(), key, data, time.Second*30)
	}
	fmt.Println("getFrom database")

	reponsePass := types.InventoriesResponse{
		Status: statusCode.Success("OK"),
		Data:   &reponseData,
	}

	return &reponsePass, nil
}

func (i inventoryServiceRedis) FavoriteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "FavoriteInventory")
	logctx.Logger(id, "id")
	preload := []string{"InventoryType", "InventoryBranch", "InventoryBrand"}
	inventory, err := i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}},
		Preload:   preload,
	})

	i.DeleteInventoryCache("inventory:shopsId:" + userInfo.UserInfo.ShopsID)
	logctx.Logger(inventory, "favoriteNaja")
	if err != nil {
		logctx.Logger(err.Error(), "[error-api] favorite Inventory")
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Error favorite inventory"),
			Data:   nil,
		}
		return &reponseError, err
	}

	if inventory.ShopsID != userInfo.UserInfo.ShopsID || inventory.ID == "" {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	err = i.inventoryRepo.UpdateInventory(id, map[string]interface{}{
		"favorite": !inventory.Favorite,
	})

	if err != nil {
		logctx.Logger(err.Error(), "[error-api] favorite Inventory")
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Error favorite inventory"),
			Data:   nil,
		}
		return &reponseError, err
	} else {
		favoriteTxt := "favorite.add"
		if inventory.Favorite {
			favoriteTxt = "favorite.delete"
		}

		reponseSuccess := types.MutationInventoryResponse{
			Status: statusCode.Success(translation.LocalizeMessage(favoriteTxt)),
			Data: &types.ResponseID{
				ID: &inventory.ID,
			},
		}
		return &reponseSuccess, nil
	}

}

func (i inventoryServiceRedis) UpsertInventory(input types.UpsertInventoryInput, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "UpsertInventory")
	logctx.Logger(input, "input")

	name := input.Name + "|" + input.InventoryBranchID + "|" + userInfo.UserInfo.ShopsID
	preload := []string{"InventoryType", "InventoryBranch", "InventoryBrand"}

	findDup, err := i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"name": name, "shops_id": userInfo.UserInfo.ShopsID}}},
		Preload:   preload,
	})

	if input.Name == "" {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("name is required"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if findDup.Name != "" && input.ID == nil {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	logctx.Logger(findDup, "findDup")
	if input.ID != nil && findDup.Name != "" && *input.ID != findDup.ID {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest(translation.LocalizeMessage("Upsert.duplicated")),
			Data:   nil,
		}
		return &reponseError, nil
	}

	findDup = model.Inventory{
		ID:          "",
		Name:        name,
		Description: input.Description,
		ShopsID:     userInfo.UserInfo.ShopsID,
		CreatedBy:   userInfo.UserInfo.UserName,
		UpdatedBy:   userInfo.UserInfo.UserName,
	}

	inventoryData := dto.MapInputToInventory(input, userInfo)
	onOkLocalT := "Upsert.success.create.inventory"
	saveFailedLocalT := "Upsert.failed.create"
	logctx.Logger(inventoryData, "InventoryData", true)
	if input.ID != nil {
		findDup.ID = *input.ID
		onOkLocalT = "Upsert.success.update.inventory"
		saveFailedLocalT = "Upsert.failed.update"
		if findDup.ShopsID != userInfo.UserInfo.ShopsID {
			reponseError := types.MutationInventoryResponse{
				Status: statusCode.Forbidden("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
	}

	logctx.Logger(inventoryData, "InventoryData", true)

	err = i.inventoryRepo.SaveInventory(inventoryData)

	if err != nil {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError(translation.LocalizeMessage(saveFailedLocalT)),
			Data:   nil,
		}
		return &reponseError, nil
	} else {
		i.DeleteInventoryCache("inventory:shopsId:" + userInfo.UserInfo.ShopsID)
		reponsePass := types.MutationInventoryResponse{
			Status: statusCode.Success(translation.LocalizeMessage(onOkLocalT)),
			Data: &types.ResponseID{
				ID: &inventoryData.ID,
			},
		}
		return &reponsePass, nil
	}

}
func (i inventoryServiceRedis) GetInventory(id *string, userInfo middlewares.UserClaims) (*types.InventoryDataResponse, error) {
	logctx := helper.LogContext(ClassName, "GetInventory")
	logctx.Logger(id, "id")
	// inventory := &model.Inventory{}

	preload := []string{"InventoryType", "InventoryBranch", "InventoryBrand"}

	inventory, _ := i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}},
		Preload:   preload,
	})

	logctx.Logger(inventory, "inventory")

	if userInfo.UserInfo.ShopsID == "" || userInfo.UserInfo.UserName == "" {
		reponseError := types.InventoryDataResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	if inventory.ShopsID != userInfo.UserInfo.ShopsID {
		reponseError := types.InventoryDataResponse{
			Status: statusCode.Forbidden("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	var wg sync.WaitGroup

	var inventoryBrand = types.InventoryBrand{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		inventoryBrand = types.InventoryBrand{
			ID:          &inventory.InventoryBrand.ID,
			Name:        strings.Split(inventory.InventoryBrand.Name, "|")[0],
			Description: inventory.InventoryBrand.Description,
			CreatedBy:   &inventory.InventoryBrand.CreatedBy,
			CreatedAt:   inventory.InventoryBrand.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryBrand.UpdatedBy,
			UpdatedAt:   inventory.InventoryBrand.UpdatedAt.UTC().Format(time.DateTime),
		}
	}()

	var inventoryBranch = types.InventoryBranch{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		inventoryBranch = types.InventoryBranch{
			ID:          &inventory.InventoryBranch.ID,
			Name:        strings.Split(inventory.InventoryBranch.Name, "|")[0],
			Description: inventory.InventoryBranch.Description,
			CreatedBy:   &inventory.InventoryBranch.CreatedBy,
			CreatedAt:   inventory.InventoryBranch.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryBranch.UpdatedBy,
			UpdatedAt:   inventory.InventoryBranch.UpdatedAt.UTC().Format(time.DateTime),
		}
	}()

	var inventoryType = types.InventoryType{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		inventoryType = types.InventoryType{
			ID:          &inventory.InventoryType.ID,
			Name:        strings.Split(inventory.InventoryType.Name, "|")[0],
			Description: inventory.InventoryType.Description,
			CreatedBy:   &inventory.InventoryType.CreatedBy,
			CreatedAt:   inventory.InventoryType.CreatedAt.UTC().Format(time.DateTime),
			UpdatedBy:   &inventory.InventoryType.UpdatedBy,
			UpdatedAt:   inventory.InventoryType.UpdatedAt.UTC().Format(time.DateTime),
		}
	}()

	wg.Wait()

	expiryDateStr := ""
	if inventory.ExpiryDate != nil {
		expiryDateStr = inventory.ExpiryDate.UTC().Format(time.DateOnly)
	}

	inventoryData := &types.Inventory{
		ID:              &inventory.ID,
		Name:            strings.Split(inventory.Name, "|")[0],
		Description:     inventory.Description,
		CreatedBy:       &inventory.CreatedBy,
		CreatedAt:       inventory.CreatedAt.UTC().Format(time.DateTime),
		UpdatedBy:       &inventory.UpdatedBy,
		UpdatedAt:       inventory.UpdatedAt.UTC().Format(time.DateTime),
		Amount:          inventory.Amount,
		Sold:            &inventory.Sold,
		Sku:             inventory.SKU,
		SerialNumber:    inventory.SerialNumber,
		Size:            inventory.Size,
		PriceMember:     inventory.PriceMember,
		Price:           inventory.Price,
		ReorderLevel:    inventory.ReorderLevel,
		ExpiryDate:      &expiryDateStr,
		InventoryBrand:  &inventoryBrand,
		InventoryBranch: &inventoryBranch,
		InventoryType:   &inventoryType,
		Favorite:        &inventory.Favorite,
	}

	reponse := &types.InventoryDataResponse{
		Data:   inventoryData,
		Status: statusCode.Success("OK"),
	}

	return reponse, nil
}

func (i inventoryServiceRedis) DeleteInventory(id string, userInfo middlewares.UserClaims) (*types.MutationInventoryResponse, error) {
	logctx := helper.LogContext(ClassName, "DeleteInventory")
	logctx.Logger(id, "id")

	preload := []string{}

	inventory, _ := i.inventoryRepo.FindFirstInventory(repositories.FindParams{
		WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": id}}},
		Preload:   preload,
	})

	if inventory.ShopsID != userInfo.UserInfo.ShopsID || inventory.ID == "" {
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.BadRequest("Unauthorized ShopId"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	err := i.inventoryRepo.UpdateInventory(id, map[string]interface{}{
		"deleted": true,
	})

	if err != nil {
		logctx.Logger(id, "[error-api] deleting Inventory")
		reponseError := types.MutationInventoryResponse{
			Status: statusCode.InternalError("Error deleting inventory"),
			Data:   nil,
		}
		return &reponseError, err
	} else {
		reponseSuccess := types.MutationInventoryResponse{
			Status: statusCode.Success(translation.LocalizeMessage("Success.inventory")),
			Data:   nil,
		}
		return &reponseSuccess, nil
	}

}

func (i inventoryServiceRedis) RecoveryHardDeleted(input types.RecoveryHardDeletedInput, userInfo middlewares.UserClaims) (*types.RecoveryHardDeletedResponse, error) {
	logctx := helper.LogContext(ClassName, "RecoveryHardDeleted")
	switch input.Type {
	case types.DeletedTypeInventory:
		{
			checkIn, err := i.inventoryRepo.FindFirstInventory(repositories.FindParams{
				WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": input.ID}}},
			})

			if err != nil {
				logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted")
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.InternalError("Error Recovery Hard Deleted"),
					Data:   nil,
				}
				return &reponseError, err
			}
			logctx.Logger(checkIn.Deleted, "checkIn")

			if checkIn.ShopsID != userInfo.UserInfo.ShopsID || checkIn.Deleted == false {
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.BadRequest("Unauthorized ShopId"),
					Data:   nil,
				}
				return &reponseError, nil
			}
			return subRecovery(&checkIn, input, userInfo, i)
		}
	case types.DeletedTypeInventoryBranch:
		{
			checkIn, err := i.inventoryRepo.FindFirstInventoryBranch(repositories.FindParams{
				WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": input.ID}}},
			})
			if err != nil {
				logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted")
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.InternalError("Error Recovery Hard Deleted"),
					Data:   nil,
				}
				return &reponseError, err
			}
			logctx.Logger(checkIn, "checkIn")

			if checkIn.ShopsID != userInfo.UserInfo.ShopsID || checkIn.Deleted == false {
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.BadRequest("Unauthorized ShopId"),
					Data:   nil,
				}
				return &reponseError, nil
			}
			return subRecovery(&checkIn, input, userInfo, i)
		}
	case types.DeletedTypeInventoryBrand:
		{
			checkIn, err := i.inventoryRepo.FindFirstInventoryBrand(repositories.FindParams{
				WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": input.ID}}},
			})

			if err != nil {
				logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted")
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.InternalError("Error Recovery Hard Deleted"),
					Data:   nil,
				}
				return &reponseError, err
			}
			logctx.Logger(checkIn, "checkIn")

			if checkIn.ShopsID != userInfo.UserInfo.ShopsID || checkIn.Deleted == false {
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.BadRequest("Unauthorized ShopId"),
					Data:   nil,
				}
				return &reponseError, nil
			}
			return subRecovery(&checkIn, input, userInfo, i)
		}
	default:
		checkIn, err := i.inventoryRepo.FindFirstInventoryType(repositories.FindParams{
			WhereArgs: []repositories.WhereArgs{{Where: map[string]interface{}{"id": input.ID}}},
		})
		if err != nil {
			logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted")
			reponseError := types.RecoveryHardDeletedResponse{
				Status: statusCode.InternalError("Error Recovery Hard Deleted"),
				Data:   nil,
			}
			return &reponseError, err
		}
		logctx.Logger(checkIn, "checkIn")

		if checkIn.ShopsID != userInfo.UserInfo.ShopsID || checkIn.Deleted == false {
			reponseError := types.RecoveryHardDeletedResponse{
				Status: statusCode.BadRequest("Unauthorized ShopId"),
				Data:   nil,
			}
			return &reponseError, nil
		}
		return subRecovery(&checkIn, input, userInfo, i)
	}
}

func subRecovery(checkIn interface{}, input types.RecoveryHardDeletedInput, userInfo middlewares.UserClaims, i inventoryServiceRedis) (*types.RecoveryHardDeletedResponse, error) {
	logctx := helper.LogContext(ClassName, "RecoveryHardDeletedSub")

	switch input.Mode {
	case types.DeletedModeDelete:
		{
			err := i.inventoryRepo.DeleteSubInventory(checkIn, input.ID)
			if err != nil {
				logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted checkIn")
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.InternalError("Error Recovery Hard Deleted checkIn"),
					Data:   nil,
				}
				return &reponseError, err
			}

			response := types.RecoveryHardDeletedResponse{
				Status: statusCode.Success(translation.LocalizeMessage("Success.trash.delete")),
				Data:   nil,
			}

			return &response, nil
		}
	default:
		{
			err := i.inventoryRepo.RecoverySubInventory(checkIn, input.ID, map[string]interface{}{"deleted": false}, userInfo.UserInfo.UserName)
			if err != nil {
				logctx.Logger(err.Error(), "[error-api] Recovery Hard Deleted checkIn")
				reponseError := types.RecoveryHardDeletedResponse{
					Status: statusCode.InternalError("Error Recovery Hard Deleted checkIn"),
					Data:   nil,
				}
				return &reponseError, err
			}

			response := types.RecoveryHardDeletedResponse{
				Status: statusCode.Success(translation.LocalizeMessage("Success.trash.recovery")),
				Data:   nil,
			}

			return &response, nil
		}
	}
}

func (i inventoryServiceRedis) GetInventoryNames(userInfo middlewares.UserClaims) (*types.InventoryNameResponse, error) {
	logctx := LogContext(ClassName, "GetInventoryNames")
	findParams := repositories.FindParams{
		WhereArgs:   []repositories.WhereArgs{{Where: map[string]interface{}{"shops_id": userInfo.UserInfo.ShopsID}}},
		SelectField: []string{"id", "name"},
	}

	inventories, _ := i.inventoryRepo.FindInventory(findParams)

	logctx.Logger(inventories, "inventories")
	inventoryName := make([]*types.InventoryName, len(inventories))

	for d, inventory := range inventories {
		inventoryName[d] = &types.InventoryName{
			ID:   &inventory.ID,
			Name: &strings.Split(inventory.Name, "|")[0],
		}
	}

	reponseSuccess := types.InventoryNameResponse{
		Status: statusCode.Success("OK"),
		Data:   inventoryName,
	}
	return &reponseSuccess, nil

}

func (i inventoryServiceRedis) GetInventoryAllDeleted(userInfo middlewares.UserClaims) (*types.DeletedInventoryResponse, error) {
	start := time.Now()
	logctx := LogContext(ClassName, "GetInventoryAllDeleted")
	logctx.Logger(userInfo, "userInfo")

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// Channels for concurrent data
	chInventory := make(chan []*types.DeletedInventoryType, 1)
	chType := make(chan []*types.DeletedInventoryType, 1)
	chBrand := make(chan []*types.DeletedInventoryType, 1)
	chBranch := make(chan []*types.DeletedInventoryType, 1)

	// Goroutine for Inventories
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chInventory)

		inventories, err := i.inventoryRepo.FindInventory(repositories.FindParams{
			WhereArgs:   []repositories.WhereArgs{{Where: map[string]interface{}{"shops_id": userInfo.UserInfo.ShopsID, "deleted": true}}},
			SelectField: []string{"id", "name"},
			OrderBy:     "updated_at desc",
		})
		if err != nil {
			errChan <- err
			return
		}

		logctx.Logger(inventories, "inventories")
		deletedInventory := make([]*types.DeletedInventoryType, len(inventories))
		for d, inventory := range inventories {
			deletedInventory[d] = &types.DeletedInventoryType{
				ID:        &inventory.ID,
				Name:      &strings.Split(inventory.Name, "|")[0],
				CreatedBy: &inventory.CreatedBy,
				UpdatedBy: &inventory.UpdatedBy,
				UpdatedAt: inventory.UpdatedAt.UTC().Format(time.DateTime),
				CreatedAt: inventory.CreatedAt.UTC().Format(time.DateTime),
			}
		}
		chInventory <- deletedInventory
	}()

	// Goroutine for Types
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chType)

		inventoryTypes, err := i.inventoryRepo.FindInventoryType(repositories.FindParams{
			WhereArgs:   []repositories.WhereArgs{{Where: map[string]interface{}{"shops_id": userInfo.UserInfo.ShopsID, "deleted": true}}},
			SelectField: []string{"id", "name"},
			OrderBy:     "updated_at desc",
		})
		if err != nil {
			errChan <- err
			return
		}

		logctx.Logger(inventoryTypes, "inventoryTypes")
		deletedInventoryType := make([]*types.DeletedInventoryType, len(inventoryTypes))
		for d, inventoryType := range inventoryTypes {
			deletedInventoryType[d] = &types.DeletedInventoryType{
				ID:        &inventoryType.ID,
				Name:      &strings.Split(inventoryType.Name, "|")[0],
				CreatedBy: &inventoryType.CreatedBy,
				UpdatedBy: &inventoryType.UpdatedBy,
				UpdatedAt: inventoryType.UpdatedAt.UTC().Format(time.DateTime),
				CreatedAt: inventoryType.CreatedAt.UTC().Format(time.DateTime),
			}
		}
		chType <- deletedInventoryType
	}()

	// Goroutine for Brands
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chBrand)

		inventoryBrands, err := i.inventoryRepo.FindInventoryBrand(repositories.FindParams{
			WhereArgs:   []repositories.WhereArgs{{Where: map[string]interface{}{"shops_id": userInfo.UserInfo.ShopsID, "deleted": true}}},
			SelectField: []string{"id", "name"},
			OrderBy:     "updated_at desc",
		})
		if err != nil {
			errChan <- err
			return
		}

		logctx.Logger(inventoryBrands, "inventoryBrands")
		deletedInventoryBrand := make([]*types.DeletedInventoryType, len(inventoryBrands))
		for d, inventoryBrand := range inventoryBrands {
			deletedInventoryBrand[d] = &types.DeletedInventoryType{
				ID:        &inventoryBrand.ID,
				Name:      &strings.Split(inventoryBrand.Name, "|")[0],
				CreatedBy: &inventoryBrand.CreatedBy,
				UpdatedBy: &inventoryBrand.UpdatedBy,
				UpdatedAt: inventoryBrand.UpdatedAt.UTC().Format(time.DateTime),
				CreatedAt: inventoryBrand.CreatedAt.UTC().Format(time.DateTime),
			}
		}
		chBrand <- deletedInventoryBrand
	}()

	// Goroutine for Branches
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chBranch)

		inventoryBranches, err := i.inventoryRepo.FindInventoryBranch(repositories.FindParams{
			WhereArgs:   []repositories.WhereArgs{{Where: map[string]interface{}{"shops_id": userInfo.UserInfo.ShopsID, "deleted": true}}},
			SelectField: []string{"id", "name"},
			OrderBy:     "updated_at desc",
		})
		if err != nil {
			errChan <- err
			return
		}
		logctx.Logger(inventoryBranches, "inventoryBranches")
		deletedInventoryBranch := make([]*types.DeletedInventoryType, len(inventoryBranches))
		for d, inventoryBranch := range inventoryBranches {
			deletedInventoryBranch[d] = &types.DeletedInventoryType{
				ID:        &inventoryBranch.ID,
				Name:      &strings.Split(inventoryBranch.Name, "|")[0],
				CreatedBy: &inventoryBranch.CreatedBy,
				UpdatedBy: &inventoryBranch.UpdatedBy,
				UpdatedAt: inventoryBranch.UpdatedAt.UTC().Format(time.DateTime),
				CreatedAt: inventoryBranch.CreatedAt.UTC().Format(time.DateTime),
			}
		}
		chBranch <- deletedInventoryBranch
	}()

	wg.Wait()
	close(errChan)

	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	response := types.DeletedInventoryResponse{
		Data: &types.DeletedInventory{
			Inventory: <-chInventory,
			Type:      <-chType,
			Brand:     <-chBrand,
			Branch:    <-chBranch,
		},
		Status: statusCode.Success("OK"),
	}

	elapsed := time.Since(start)
	log.Println("Execution time:", elapsed)
	return &response, nil
}
