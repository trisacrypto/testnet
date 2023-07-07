package openvasp

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The Customer struct binds to JSON sent to
// the Register endpoint when executing
// registering a contact and is used to store
// contacts in the database
type Customer struct {
	gorm.Model
	CustomerID    uuid.UUID    `gorm:"uniqueIndex;size:255;column:customer_id;not null"`
	Name          string       `gorm:"column:name;not null"`
	AssetType     VirtualAsset `gorm:"column:asset_type;not null"`
	WalletAddress string       `gorm:"column:wallet_address;not null"`
	TravelAddress string       `gorm:"column:travel_address;not null"`
}

// The Payload struct binds to JSON sent to
// the Transfer endpoint
type Payload struct {
	IVMS101   string
	AssetType VirtualAsset
	Amount    float64
	Callback  string
	Reject    bool
}

type VirtualAsset uint16

const (
	UnknownAsset VirtualAsset = iota
	Bitcoin
	Tether
	Ethereum
	Litecoin
	XRP
	BitcoinCash
	Tezos
	EOS
)

// The Transfer struct is constructed from
// Payload data sent to the transfer endpoint
// and is used to store transfers in the database
type Transfer struct {
	gorm.Model
	TransferID     uuid.UUID      `gorm:"uniqueIndex;size:255;column:transfer_id;not null"`
	Status         TransferStatus `gorm:"column:status;not null"`
	OriginatorVasp string         `gorm:"column:originator_vasp;not null"`
	Originator     string         `gorm:"column:originator;not null"`
	Beneficiary    string         `gorm:"column:beneficiary;not null"`
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

// TransferApproval struct binds to JSON sent to
// the TransferInquiry endpoint when executing
// the callback provided by a Transfer call
type TransferReply struct {
	Approved *TransferApproval
	Rejected string
}

type TransferApproval struct {
	Address  string
	Callback string
}

// TransferConfirmation struct binds to JSON sent to
// the TransferConfirmation endpoint
type TransferConfirmation struct {
	TxId     string
	Canceled string
}

// Wraps a GORM database and contains
// handlers for the Gin endpoints
type server struct {
	db *gorm.DB
}

// Create a new Server object containing a GORM database
func New(dsn string) (newServer *server, err error) {
	newServer = &server{}
	if newServer.db, err = openDB(dsn); err != nil {
		return nil, err
	}
	return newServer, nil
}

// Opens a new GORM database and performs the migration
func openDB(dsn string) (db *gorm.DB, err error) {
	if db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&Customer{}, &Transfer{}); err != nil {
		return nil, err
	}
	return db, nil
}
