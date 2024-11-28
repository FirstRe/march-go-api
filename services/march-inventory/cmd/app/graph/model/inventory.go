package model

import (
	"time"
)

type StatusFile string

const (
	PASS StatusFile = "PASS"
	FAIL StatusFile = "FAIL"
)

type InventoryBrand struct {
	ID          string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name        string      `gorm:"unique" json:"name"`
	ShopsID     string      `gorm:"type:varchar(40)" json:"shopsId"`
	Description *string     `gorm:"type:varchar(300)" json:"description,omitempty"`
	Deleted     bool        `gorm:"default:false" json:"deleted"`
	CreatedBy   string      `json:"createdBy"`
	UpdatedBy   string      `json:"updatedBy"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt   time.Time   `gorm:"default:now()" json:"createdAt"`
	Inventory   []Inventory `gorm:"foreignKey:InventoryBrandID" json:"inventory"`
}

type InventoryType struct {
	ID          string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name        string      `gorm:"unique" json:"name"`
	ShopsID     string      `gorm:"type:varchar(40)" json:"shopsId"`
	Description *string     `gorm:"type:varchar(300)" json:"description,omitempty"`
	Deleted     bool        `gorm:"default:false" json:"deleted"`
	CreatedBy   string      `json:"createdBy"`
	UpdatedBy   string      `json:"updatedBy"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt   time.Time   `gorm:"default:now()" json:"createdAt"`
	Inventory   []Inventory `gorm:"foreignKey:InventoryTypeID" json:"inventory"`
}

type InventoryBranch struct {
	ID          string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name        string      `gorm:"unique" json:"name"`
	ShopsID     string      `gorm:"type:varchar(40)" json:"shopsId"`
	Description *string     `gorm:"type:varchar(300)" json:"description,omitempty"`
	Deleted     bool        `gorm:"default:false" json:"deleted"`
	CreatedBy   string      `json:"createdBy"`
	UpdatedBy   string      `json:"updatedBy"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt   time.Time   `gorm:"default:now()" json:"createdAt"`
	Inventory   []Inventory `gorm:"foreignKey:InventoryBranchID" json:"inventory"`
}

type Inventory struct {
	ID                string          `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name              string          `gorm:"unique" json:"name"`
	InventoryTypeID   string          `gorm:"type:varchar(40)" json:"inventoryTypeId"`
	InventoryType     InventoryType   `gorm:"foreignKey:InventoryTypeID;references:ID" json:"inventoryType"`
	InventoryBrandID  string          `gorm:"type:varchar(40)" json:"inventoryBrandId"`
	InventoryBrand    InventoryBrand  `gorm:"foreignKey:InventoryBrandID;references:ID" json:"inventoryBrand"`
	InventoryBranchID string          `gorm:"type:varchar(40)" json:"inventoryBranchId"`
	InventoryBranch   InventoryBranch `gorm:"foreignKey:InventoryBranchID;references:ID" json:"inventoryBranch"`
	Amount            int             `gorm:"default:0" json:"amount"`
	Price             int             `gorm:"default:0" json:"price"`
	PriceMember       *int            `gorm:"default:0" json:"priceMember,omitempty"`
	Size              *string         `json:"size,omitempty"`
	SKU               *string         `json:"sku,omitempty"`
	SerialNumber      *string         `json:"serialNumber,omitempty"`
	ReorderLevel      *int            `json:"reorderLevel,omitempty"`
	Sold              int             `gorm:"default:0" json:"sold"`
	ExpiryDate        *time.Time      `json:"expiryDate,omitempty"`
	ShopsID           string         `gorm:"type:varchar(40)" json:"shopsId,omitempty"`
	Description       *string         `gorm:"type:varchar(320)" json:"description,omitempty"`
	Deleted           bool            `gorm:"default:false" json:"deleted"`
	CreatedBy         string          `json:"createdBy"`
	UpdatedBy         string          `json:"updatedBy"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt         time.Time       `gorm:"default:now()" json:"createdAt"`
	Favorite          bool            `gorm:"default:false" json:"favorite"`
	CSVID             *string         `json:"csvId,omitempty"`
	InventoryFileID   *string         `gorm:"type:varchar(40)" json:"inventoryFileId,omitempty"`
	InventoryFile     *InventoryFile  `gorm:"foreignKey:InventoryFileID;references:ID" json:"inventoryFile,omitempty"`
}

type InventoryFile struct {
	ID        string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name      string      `gorm:"unique" json:"name"`
	ShopsID   *string     `gorm:"type:varchar(40)" json:"shopsId,omitempty"`
	Status    StatusFile  `gorm:"default:PASS" json:"status"`
	CreatedBy string      `json:"createdBy"`
	UpdatedBy string      `json:"updatedBy"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt time.Time   `gorm:"default:now()" json:"createdAt"`
	Inventory []Inventory `gorm:"foreignKey:InventoryFileID" json:"inventory"`
}
