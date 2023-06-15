package openvasp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Customer struct {
	ID            uuid.UUID
	Name          string
	AssetType     VirtualAsset
	WalletAddress string
	TravelAddress string
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

// Register a new customer. This will take in a customer ID
// (and will generate one if it is not provided), customer name
// and Asset type, and will then generate a LNURL associated with
// this customer.
func Register(c *gin.Context) {
	var newCustomer *Customer

	var err error
	if err = c.BindJSON(newCustomer); err != nil {
		return
	}

	if err = validateCustomer(newCustomer); err != nil {
		return
	}

	//TODO: generate LNURL TravelAddress
	c.IndentedJSON(http.StatusCreated, newCustomer)
}

// Validate that the Registration JSON is valid
func validateCustomer(customer *Customer) (err error) {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}

	if customer.AssetType == UnknownAsset {
		return errors.New("asset must be set")
	}

	if customer.WalletAddress == "" {
		return errors.New("wallet Address must be set")
	}
	return nil
}

//TODO: implement
func Transfer(c *gin.Context) {}
