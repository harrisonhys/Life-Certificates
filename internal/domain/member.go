package domain

import "time"

// Member represents an individual enrolled in the programme.
type Member struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	NIK          string    `gorm:"size:20;uniqueIndex" json:"nik"`
	NomorPeserta string    `gorm:"size:50;uniqueIndex" json:"nomor_peserta"`
	BirthDate    time.Time `gorm:"type:date" json:"birth_date"`
	FullName     string    `gorm:"size:150;column:fullname" json:"fullname"`
	Address      string    `gorm:"size:255" json:"address"`
	City         string    `gorm:"size:100" json:"city"`
	Province     string    `gorm:"size:100" json:"province"`
	PhoneNumber  string    `gorm:"size:30;column:phone_number" json:"phone_number"`
	Email        string    `gorm:"size:120" json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName keeps the table naming explicit.
func (Member) TableName() string {
	return "members"
}
