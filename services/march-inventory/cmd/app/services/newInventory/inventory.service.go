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
	inventory, err := i.inventoryRepo.FindFirstInventory(map[string]interface{}{"id": id}, preload)

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
	findDup, err := i.inventoryRepo.FindFirstInventory(map[string]interface{}{"name": name, "shops_id": userInfo.UserInfo.ShopsID}, preload)

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

	inventory, _ := i.inventoryRepo.FindFirstInventory(map[string]interface{}{"id": id}, preload)

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
	logctx := helper.LogContext(ClassName, "GetInventory")
	logctx.Logger(id, "id")

	preload := []string{}

	inventory, _ := i.inventoryRepo.FindFirstInventory(map[string]interface{}{"id": id}, preload)

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
