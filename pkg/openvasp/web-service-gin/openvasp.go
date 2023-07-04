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
	IVMS101   *trisa.IdentityPayload `binding:"required"`
	AssetType VirtualAsset           `binding:"required"`
	Amount    float64                `binding:"required"`
	Callback  string                 `binding:"required"`
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
	Approved *TransferApproval
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
	router.GET("/getlnurl/:id", s.GetLNURL)
	router.POST("/transfer", s.Transfer)
	router.GET("/gettransfer/:id", s.GetTransfer)
	router.POST("/inquiryResolution/:id", s.InquiryResolution)
	router.POST("/transferConfirmation/:id", s.TransferConfirmation)
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

func (s *server) GetLNURL(c *gin.Context) {
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

func validatePayload(payload *Payload) (err error) {
	if payload.IVMS101 == nil {
		return errors.New("ivms101 payload must be set")
	}

	if payload.AssetType == UnknownAsset {
		return errors.New("asset type must be set")
	}

	if payload.Amount == 0 {
		return errors.New("transfer amount must be set")
	}

	if payload.Callback == "" {
		return errors.New("callback must be set")
	}
	return nil
}

func (s *server) GetTransfer(c *gin.Context) {
	var err error
	var TransferID uuid.UUID
	if TransferID, err = uuid.Parse(c.Param("id")); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not parse id": err.Error()})
		return
	}

	var transfer Transfer
	if db := s.db.Where("transfer_id = ?", TransferID).First(&TransferID); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not find transfer": db.Error})
	}
	c.IndentedJSON(http.StatusFound, &transfer)
}

func (s *server) InquiryResolution(c *gin.Context) {
	fmt.Println(c.Request)

	var err error
	var reply TransferReply
	if err = c.BindJSON(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	if err = validateReply(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid resolution": err.Error()})
		return
	}

	transferID := c.Param("id")
	if reply.Approved != nil {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
		}
	} else if reply.Rejected != "" {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not reject transfer": db.Error})
		}
	}

	var transfer *Transfer
	if db := s.db.Where("customer_id = ?", transferID).First(&transfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not find updated transfer": db.Error})
	}
}

func validateReply(reply *TransferReply) error {
	if reply.Approved == nil && reply.Rejected == "" {
		return errors.New("reply must either be approved or rejected")
	}
	if reply.Approved != nil && reply.Rejected != "" {
		return errors.New("reply must either be approved or rejected")
	}
	return nil
}

func (s *server) TransferConfirmation(c *gin.Context) {
	fmt.Println(c.Request)

	var err error
	var confirmation TransferConfirmation
	if err = c.BindJSON(&confirmation); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}
}
