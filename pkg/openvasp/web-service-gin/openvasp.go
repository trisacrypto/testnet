package openvasp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	trisa "github.com/trisacrypto/trisa/pkg/ivms101"
	"google.golang.org/protobuf/encoding/protojson"
)

const travelURLTemplate = "http://localhost:4435/transfer/%s?tag=travelRuleInquiry"

// Serves the Gin server on the provided address, creates a
// Postgress database on the provided DSN and creates the
// Gin endpoint handlers
func Serve(address, gormDSN string) (err error) {
	var s *server
	if s, err = New(gormDSN); err != nil {
		return err
	}

	router := gin.Default()
	router.POST("/register", s.Register)
	router.GET("/listusers", s.ListUsers)
	router.GET("/gettraveladdress/:id", s.GetTravelAddress)
	router.POST("/transfer/:id", s.Transfer)
	router.GET("/gettransfer/:id", s.GetTransfer)
	router.POST("/inquiryresolution/:id", s.InquiryResolution)
	router.Run(address)
	return nil
}

// Register a new customer. This will take in a customer ID
// (and will generate one if it is not provided), customer name
// and Asset type, and will then generate a LNURL associated with
// this customer.
func (s *server) Register(c *gin.Context) {
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
	if newCustomer.TravelAddress, err = lnurl.LNURLEncode(travelAddress); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": err.Error()})
		return
	}

	if db := s.db.Create(&newCustomer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": db.Error})
		return
	}
	c.IndentedJSON(http.StatusCreated, &newCustomer)
}

// Helper function to ensure that the JSON provided to the register
// endpoint is valid
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

// The ListUsers endpoint will return a list of the registered users'
// ID, name and LNURL encoded TravelAddress
func (s *server) ListUsers(c *gin.Context) {
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
	c.IndentedJSON(http.StatusOK, &users)
}

type user struct {
	CustomerID uuid.UUID
	Name       string
	LNURL      string
}

// The GetTravelAddress endpoint returns the ID, name and
// LNURL encoded TravelAddress of the registered user associated
// with the provided CustomerID
func (s *server) GetTravelAddress(c *gin.Context) {
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
	c.IndentedJSON(http.StatusOK, &foundUser)
}

//
func (s *server) Transfer(c *gin.Context) {
	var err error
	var newPayload Payload
	if err = c.BindJSON(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}
	newPayload.IVMS101 = strings.ReplaceAll(newPayload.IVMS101, `*`, `"`)
	newPayload.IVMS101 = strings.ReplaceAll(newPayload.IVMS101, "+", "\n")

	ivms101 := &trisa.IdentityPayload{}
	jsonpb := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	if err = jsonpb.Unmarshal([]byte(newPayload.IVMS101), ivms101); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not unmarshal ivms101": err.Error()})
		return
	}

	if err = validatePayload(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid payload provided": err.Error()})
		return
	}

	newTransfer := Transfer{
		TransferID:     uuid.New(),
		Status:         Pending,
		OriginatorVasp: originatorVasp(ivms101),
		Originator:     originatorName(ivms101),
		Beneficiary:    beneficiaryName(ivms101),
		AssetType:      newPayload.AssetType,
		Amount:         newPayload.Amount,
		Created:        time.Now(),
	}

	if db := s.db.Create(&newTransfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not create transfer": db.Error})
		return
	}
	c.IndentedJSON(http.StatusCreated, &newTransfer)
}

// Helper function to ensure that the JSON provided to the transfer
// endpoint is valid
func validatePayload(payload *Payload) (err error) {
	if payload.IVMS101 == "" {
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

// Helper function to extract the originator name identifier from
// a trisa Identity Payload
func originatorName(payload *trisa.IdentityPayload) string {
	originator := payload.Originator
	nameIds := originator.GetOriginatorPersons()[0].GetNaturalPerson().Name.NameIdentifiers[0]
	return fmt.Sprintf("%s %s", nameIds.PrimaryIdentifier, nameIds.SecondaryIdentifier)
}

// Helper function to extract the beneficiary name identifier from
// a trisa Identity Payload
func beneficiaryName(payload *trisa.IdentityPayload) string {
	beneficiary := payload.Beneficiary
	nameIds := beneficiary.GetBeneficiaryPersons()[0].GetNaturalPerson().Name.NameIdentifiers[0]
	return fmt.Sprintf("%s %s", nameIds.PrimaryIdentifier, nameIds.SecondaryIdentifier)
}

// Helper function to extract the originating vasp name from
// a trisa Identity Payload
func originatorVasp(payload *trisa.IdentityPayload) string {
	originatingVasp := payload.OriginatingVasp
	vaspName := originatingVasp.GetOriginatingVasp().GetLegalPerson().Name.NameIdentifiers[0].LegalPersonName
	return vaspName
}

// The GetTransfer endpoint returns the details of the
// transfer identified by the specified transfer_id
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
	c.IndentedJSON(http.StatusOK, &transfer)
}

//
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
	if reply.Approved != "" {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
		}
		c.IndentedJSON(http.StatusAccepted, &TransferConfirmation{TxId: transferID})
	} else if reply.Rejected != "" {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not reject transfer": db.Error})
		}
		c.IndentedJSON(http.StatusAccepted, &TransferConfirmation{Canceled: "The Transaction has been Canceled"})
	}

	var transfer *Transfer
	if db := s.db.Where("customer_id = ?", transferID).First(&transfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not find updated transfer": db.Error})
	}
}

// Helper function to ensure that the JSON provided to the
// inquiryresolution endpoint is valid
func validateReply(reply *TransferReply) error {
	if reply.Callback == "" {
		return errors.New("reply must have a callback")
	}

	err := errors.New("reply must either be approved or rejected")
	if reply.Approved == "" && reply.Rejected == "" {
		return err
	}
	if reply.Approved != "" && reply.Rejected != "" {
		return err
	}
	return nil
}

//
func (s *server) TransferConfirmation(c *gin.Context) {
	fmt.Println(c.Request)

	var err error
	var confirmation TransferConfirmation
	if err = c.BindJSON(&confirmation); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	if err = validateConfirmation(&confirmation); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid resolution": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Helper function to ensure that the JSON provided to the
// transferconfirmation endpoint is valid
func validateConfirmation(confirmation *TransferConfirmation) error {
	err := errors.New("confirmation must either have transfer ID or cancelation")
	if confirmation.TxId == "" && confirmation.Canceled == "" {
		return err
	}
	if confirmation.TxId != "" && confirmation.Canceled != "" {
		return err
	}
	return nil
}
