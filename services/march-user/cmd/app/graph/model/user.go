package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string    `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	GroupID      string    `gorm:"type:varchar(40);not null" json:"groupId"`
	Group        Group     `gorm:"foreignKey:GroupID;references:ID" json:"group"`
	ShopsID      string    `gorm:"type:varchar(40);not null" json:"shopsId"`
	Shop         Shop      `gorm:"foreignKey:ShopsID;references:ID" json:"shop"`
	Username     string    `gorm:"not null" json:"username"`
	Password     *string   `gorm:"null" json:"password"`
	IsSuperAdmin *bool     `gorm:"default:false" json:"isSuperAdmin"`
	Email        *string   `gorm:"unique" json:"email"`
	IsRegistered *bool     `gorm:"default:true" json:"isRegistered"`
	Picture      *string   `gorm:"null" json:"picture"`
	RefreshToken *string   `gorm:"type:varchar(300);null" json:"refreshToken"`
	Deleted      bool      `gorm:"default:false" json:"deleted"`
	DeviceID     *string   `gorm:"null" json:"deviceId"`
	CreatedBy    string    `gorm:"not null" json:"createdBy"`
	UpdatedBy    string    `gorm:"not null" json:"updatedBy"`
	CreatedAt    time.Time `gorm:"default:now()" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type Shop struct {
	ID          string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name        string      `gorm:"unique;not null" json:"name"`
	Description *string     `gorm:"null" json:"description"`
	CreatedBy   string      `gorm:"not null" json:"createdBy"`
	UpdatedBy   string      `gorm:"not null" json:"updatedBy"`
	CreatedAt   time.Time   `gorm:"default:now()" json:"createdAt"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	Users       []User      `gorm:"foreignKey:ShopsID" json:"users"`
	Groups      []Group     `gorm:"foreignKey:ShopsID" json:"groups"`
	GroupTasks  []GroupTask `gorm:"foreignKey:ShopsID" json:"groupTasks"`
}

type GroupFunction struct {
	ID         string   `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name       string   `gorm:"unique;not null" json:"name"`
	FunctionID string   `gorm:"type:uuid;not null" json:"functionId"`
	Function   Function `gorm:"foreignKey:FunctionID;references:ID" json:"function"`
	GroupID    string   `gorm:"type:uuid;not null" json:"groupId"`
	Group      Group    `gorm:"foreignKey:GroupID;references:ID" json:"group"`
	Create     bool     `gorm:"default:false" json:"create"`
	View       bool     `gorm:"default:false" json:"view"`
	Update     bool     `gorm:"default:false" json:"update"`
}

type Function struct {
	ID             string          `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name           string          `gorm:"unique;not null" json:"name"`
	GroupFunctions []GroupFunction `gorm:"foreignKey:FunctionID" json:"groupFunctions"`
	Tasks          []Task          `gorm:"foreignKey:FunctionID" json:"tasks"`
}

type Group struct {
	ID             string          `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name           string          `gorm:"unique;not null" json:"name"`
	ShopsID        string          `gorm:"type:varchar(40);not null" json:"shopsId"`
	Shop           Shop            `gorm:"foreignKey:ShopsID;references:ID" json:"shop"`
	GroupFunctions []GroupFunction `gorm:"foreignKey:GroupID" json:"groupFunctions"`
	Users          []User          `gorm:"foreignKey:GroupID" json:"users"`
	GroupTasks     []GroupTask     `gorm:"foreignKey:GroupID" json:"groupTasks"`
}

type Task struct {
	ID          string      `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name        string      `gorm:"unique;not null" json:"name"`
	FunctionID  string      `gorm:"type:varchar(40);not null" json:"functionId"`
	Description *string     `gorm:"null" json:"description"`
	Function    Function    `gorm:"foreignKey:FunctionID;references:ID" json:"function"`
	GroupTasks  []GroupTask `gorm:"foreignKey:TaskID" json:"groupTasks"`
}

type GroupTask struct {
	ID        string    `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name      string    `gorm:"unique;not null" json:"name"`
	GroupID   string    `gorm:"type:uuid;not null" json:"groupId"`
	Group     Group     `gorm:"foreignKey:GroupID;references:ID" json:"group"`
	TaskID    string    `gorm:"type:uuid;not null" json:"taskId"`
	Task      Task      `gorm:"foreignKey:TaskID;references:ID" json:"task"`
	ShopsID   string    `gorm:"type:varchar(40);not null" json:"shopsId"`
	Shop      Shop      `gorm:"foreignKey:ShopsID;references:ID" json:"shop"`
	CreatedBy string    `gorm:"not null" json:"createdBy"`
	UpdatedBy string    `gorm:"not null" json:"updatedBy"`
	CreatedAt time.Time `gorm:"default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (base *User) BeforeCreate(scope *gorm.DB) (err error) {
	uuid := uuid.New()
	base.ID = uuid.String()
	return
}
