package openvasp

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Transfer struct {
	gorm.Model
	TransferID     uuid.UUID      `gorm:"uniqueIndex;size:255;column:transfer_id;not null"`
	Status         TransferStatus `gorm:"column:status;not null"`
	OriginatorVasp string         `gorm:"column:originator_vasp;not null"`
	Originator     string         `gorm:"column:originator;not null"`
	Beneficiaary   string         `gorm:"column:beneficiary;not null"`
	AssetType      VirtualAsset   `gorm:"column:asset_type;not null"`
	Amount         float64        `gorm:"column:amount;not null"`
	Created        time.Time      `gorm:"column:created;not null"`
}

type TransferStatus uint16

const (
	UnknownStatus TransferStatus = iota
	Pending
	Approved
	Rejected
)

type Customer struct {
	gorm.Model
	CustomerID    uuid.UUID    `gorm:"uniqueIndex;size:255;column:customer_id;not null"`
	Name          string       `gorm:"column:name;not null"`
	AssetType     VirtualAsset `gorm:"column:asset_type;not null"`
	WalletAddress string       `gorm:"column:wallet_address;not null"`
	TravelAddress string       `gorm:"column:travel_address;not null"`
}

type server struct {
	db *gorm.DB
}

func New(dsn string) (newServer *server, err error) {
	newServer = &server{}
	if newServer.db, err = openDB(dsn); err != nil {
		return nil, err
	}
	return newServer, nil
}

func openDB(dsn string) (db *gorm.DB, err error) {
	if db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&Customer{}); err != nil {
		return nil, err
	}
	return db, nil
}
