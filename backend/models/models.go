// models/models.go
package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	Company      Company   `gorm:"foreignKey:CompanyID"`
	CompanyID    uint
	Webhooks     []Webhook `gorm:"foreignKey:UserID"`
}

type Company struct {
	gorm.Model
	Name    string  `gorm:"unique;not null"`
	Users   []User  `gorm:"foreignKey:CompanyID"`
	Sensors []Sensor `gorm:"foreignKey:CompanyID"`
}

type Sensor struct {
    gorm.Model
	ID uint  `gorm:"primaryKey" json:"id"` // Proper JSON tag
    Name        string  `gorm:"not null" json:"name"`
    Latitude    float64 `gorm:"not null" json:"latitude"`
    Longitude   float64 `gorm:"not null" json:"longitude"`
    Status      string  `gorm:"not null" json:"status"` // Proper JSON tag
    CompanyID   uint    `gorm:"not null" json:"company_id"`
    Description string  `json:"description"`
}

type Webhook struct {
	gorm.Model
	URL    string `gorm:"not null"`
	UserID uint   `gorm:"not null"`
}
