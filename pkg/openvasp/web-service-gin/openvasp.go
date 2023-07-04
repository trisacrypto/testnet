package openvasp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	trisa "github.com/trisacrypto/trisa/pkg/ivms101"
	lnurl "github.com/xplorfin/lnurlauth"
)

type Payload struct {
	IVMS101   trisa.IdentityPayload `binding:"required"`
	AssetType VirtualAsset          `binding:"required"`
	Amount    float64               `binding:"required"`
	Callback  string                `binding:"required"`
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

type TransferReply struct {
	Approved TransferApproval
	Rejected string
}

type TransferApproval struct {
	Approved string
	Callback string
}

type TransferConfirmation struct {
	TxId     uuid.UUID
	Canceled string
}

const travelURLTemplate = "https://test.net/transfer/%s?tag=travelRuleInquiry"

func (p *Payload) OriginatorName() string {
	originator := p.IVMS101.Originator
	nameIds := originator.GetOriginatorPersons()[0].GetNaturalPerson().Name.NameIdentifiers[0]
	return fmt.Sprintf("%s %s", nameIds.PrimaryIdentifier, nameIds.SecondaryIdentifier)
}

func (p *Payload) BeneficiaryName() string {
	beneficiary := p.IVMS101.Beneficiary
	nameIds := beneficiary.GetBeneficiaryPersons()[0].GetNaturalPerson().Name.NameIdentifiers[0]
	return fmt.Sprintf("%s %s", nameIds.PrimaryIdentifier, nameIds.SecondaryIdentifier)
}

func Serve(address, dsn string) (err error) {
	var s *server
	if s, err = New(dsn); err != nil {
		return err
	}

	router := gin.Default()
	router.POST("/register", s.Register)
	router.GET("/listusers", s.listUsers)
	router.GET("/lnurl/:id", s.getLNURL)
	router.POST("/transfer", s.Transfer)
	router.POST("/inquiryResolution", s.InquiryResolution)
	router.POST("/transferConfirmation", s.TransferConfirmation)
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
	fmt.Println(c.Request)

	var err error
	var newCustomer Customer
	if err = c.BindJSON(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	if err = validateCustomer(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid customer provided": err.Error()})
		return
	}

	travelAddress := fmt.Sprintf(travelURLTemplate, newCustomer.WalletAddress)
	if newCustomer.TravelAddress, _, err = lnurl.GenerateLnUrl(travelAddress); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": err.Error()})
		return
	}

	if db := s.db.Create(&newCustomer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": db.Error})
		return
	}
	c.IndentedJSON(http.StatusCreated, &newCustomer)
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

func (s *server) listUsers(c *gin.Context) {
	var users []user
	var customers []Customer
	s.db.Find(&customers)
	for _, customer := range customers {
		newUser := user{
			CustomerID: customer.CustomerID,
			Name:       customer.Name,
			LNURL:      customer.TravelAddress,
		}
		users = append(users, newUser)
	}
	c.IndentedJSON(http.StatusFound, &users)
}

type user struct {
	CustomerID uuid.UUID
	Name       string
	LNURL      string
}

func (s *server) getLNURL(c *gin.Context) {
	var err error
	var customerID uuid.UUID
	if customerID, err = uuid.Parse(c.Param("id")); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not parse id": err.Error()})
		return
	}

	var customer Customer
	s.db.Where("customer_id = ?", customerID).First(&customer)
	foundUser := user{
		CustomerID: customer.CustomerID,
		Name:       customer.Name,
		LNURL:      customer.TravelAddress,
	}
	c.IndentedJSON(http.StatusFound, &foundUser)
}

func (s *server) Transfer(c *gin.Context) {
	fmt.Println(c.Request)

	var err error
	var newPayload Payload
	if err = c.BindJSON(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	if err = validatePayload(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid payload provided": err.Error()})
		return
	}

	newTransfer := &Transfer{
		TransferID:     uuid.New(),
		Status:         Pending,
		Originator:     newPayload.OriginatorName(),
		Beneficiary:    newPayload.BeneficiaryName(),
		OriginatorVasp: newPayload.OriginatorName(),
		AssetType:      newPayload.AssetType,
		Amount:         newPayload.Amount,
		Created:        time.Now(),
	}

	if db := s.db.Create(&newTransfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": db.Error})
		return
	}
	c.IndentedJSON(http.StatusCreated, &newTransfer)
}

//TODO: implement
func validatePayload(payload *Payload) (err error) {
	return nil
}

//TODO: implement
func (s *server) InquiryResolution(c *gin.Context) {}

//TODO: implement
func (s *server) TransferConfirmation(c *gin.Context) {}
