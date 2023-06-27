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

const travelURLTemplate = "https://test.net/transfer/%s?tag=travelRuleInquiry"

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

func Serve(address, dsn string) (err error) {
	var s *server
	if s, err = New(dsn); err != nil {
		return err
	}

	router := gin.Default()
	router.POST("/register", s.Register)
	router.POST("/transfer", s.Transfer)
	router.Run(address)
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
func (s *server) Register(c *gin.Context) {
	var err error
	var newCustomer Customer
	if err = c.BindJSON(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"JSON Binding error": err.Error()})
		return
	}

	if err = validateCustomer(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"validation error": err.Error()})
		return
	}

	travelAddress := fmt.Sprintf(travelURLTemplate, newCustomer.WalletAddress)
	if newCustomer.TravelAddress, _, err = lnurl.GenerateLnUrl(travelAddress); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"LNURL generation error": err.Error()})
		return
	}

	s.db.Create(&newCustomer)
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

//TODO: implement
func (s *server) Transfer(c *gin.Context) {}
