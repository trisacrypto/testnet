package openvasp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	lnurl "github.com/xplorfin/lnurlauth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	gorm *gorm.DB
}

type Customer struct {
	gorm.Model
	CustomerID    uuid.UUID    `gorm:"uniqueIndex;size:255;column:customer_id;not null"`
	Name          string       `gorm:"column:name;null"`
	AssetType     VirtualAsset `gorm:"column:asset_type;null"`
	WalletAddress string       `gorm:"column:wallet_address;null"`
	TravelAddress string       `gorm:"column:travel_address;null"`
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

var db *DB

const travelURLTemplate = "https://test.net/transfer/%s?tag=travelRuleInquiry"

func Serve(address, dns string) (err error) {
	if db, err = OpenDB(dns); err != nil {
		return err
	}

	router := gin.Default()
	router.POST("/register", Register)
	router.POST("/transfer", Transfer)
	router.Run(fmt.Sprintf("localhost%s", address))
	return nil
}

// Register a new customer. This will take in a customer ID
// (and will generate one if it is not provided), customer name
// and Asset type, and will then generate a LNURL associated with
// this customer.
/*
Example command:
	curl http://localhost:4435/register \
			--include \
			--header "Content-Type: application/json" \
			--request "POST" --data '{"name":"Tildred Milcot", "assettype": 3, "walletaddress": "926ca69a-6c22-42e6-9105-11ab5de1237b"}'
*/
func Register(c *gin.Context) {
	var err error
	var newCustomer Customer
	if err = c.BindJSON(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = validateCustomer(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	travelAddress := fmt.Sprintf(travelURLTemplate, newCustomer.WalletAddress)
	if newCustomer.TravelAddress, _, err = lnurl.GenerateLnUrl(travelAddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.gorm.Create(&newCustomer)
	c.IndentedJSON(http.StatusCreated, newCustomer)
}

// Validate that the Registration JSON is valid
func validateCustomer(customer *Customer) (err error) {
	if customer.CustomerID == uuid.Nil {
		customer.CustomerID = uuid.New()
	}

	if customer.AssetType == UnknownAsset {
		return errors.New("asset must be set")
	}

	if customer.Name == "" {
		return errors.New("customer name must be set")
	}

	if customer.WalletAddress == "" {
		return errors.New("wallet Address must be set")
	}
	return nil
}

func OpenDB(dns string) (db *DB, err error) {
	db = &DB{}
	if db.gorm, err = gorm.Open(postgres.Open(dns), &gorm.Config{}); err == nil {
		return db, nil
	}

	if err = db.gorm.AutoMigrate(&Customer{}); err != nil {
		return nil, err
	}
	return db, nil
}

//TODO: implement
func Transfer(c *gin.Context) {}
