package openvasp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type customer struct {
	ID            uuid.UUID
	Name          string
	AssetType     VirtualAsset
	WalletAddress string
	TravelAddress string
}

type VirtualAsset int16

const (
	Bitcoin     VirtualAsset = 1
	Tether      VirtualAsset = 2
	Ethereum    VirtualAsset = 3
	Litecoin    VirtualAsset = 4
	XRP         VirtualAsset = 5
	BitcoinCash VirtualAsset = 6
	Tezos       VirtualAsset = 7
	EOS         VirtualAsset = 8
)

// Register a new customer. This will take in a customer ID
// (and will generate one if it is not provided), customer name
// and Asset type, and will then generate a LNURL associated with
// this customer.
func Register(c *gin.Context) {
	var newCustomer *customer

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
func validateCustomer(customer *customer) (err error) {
	if customer.ID == uuid.Nil {
		if customer.ID, err = uuid.NewRandom(); err != nil {
			return err
		}
	}

	if customer.AssetType == 0 {
		return errors.New("Asset must be set")
	}

	if customer.WalletAddress == "" {
		return errors.New("Wallet Address must be set")
	}
	return nil
}

//TODO: implement
func Transfer(c *gin.Context) {}
