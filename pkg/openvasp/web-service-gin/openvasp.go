package openvasp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	trisa "github.com/trisacrypto/trisa/pkg/ivms101"
	"google.golang.org/protobuf/encoding/protojson"
)

const travelURLTemplate = "%s/transfer/%s?tag=travelRuleInquiry"
const confirmationURLTemplate = "%s/transferconfirmation/%s"

// Serves the Gin server on the provided address, creates a
// Postgress database on the provided DSN and creates the
// Gin endpoint handlers
func Serve(address, callbackURL, gormDSN string) (err error) {
	var s *server
	if s, err = New(gormDSN); err != nil {
		return err
	}
	s.callbackURL = callbackURL

	router := gin.Default()
	router.POST("/register", s.Register)
	router.GET("/listusers", s.ListUsers)
	router.GET("/gettraveladdress/:id", s.GetTravelAddress)
	router.POST("/originatortransfer/:lnurl", s.OriginatorTransfer)
	router.POST("/transfer/:id", s.Transfer)
	router.GET("/listtransfers", s.ListTransfers)
	router.GET("/gettransfer/:id", s.GetTransfer)
	router.POST("/originatorconfirmation/:id", s.OriginatorConfirmation)
	router.POST("/transferconfirmation/:id", s.TransferConfirmation)
	router.Run(address)
	return nil
}

// Register a new customer. This will take in a customer ID
// (and will generate one if it is not provided), customer name
// and Asset type, and will then generate a LNURL associated with
// this customer.
func (s *server) Register(c *gin.Context) {
	// Bind the request JSON to a Customer struct
	var err error
	var newCustomer Customer
	if err = c.BindJSON(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	// Validate the received Customer struct
	if err = validateCustomer(&newCustomer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid customer provided": err.Error()})
		return
	}

	// Encode the new customer's LNURL
	travelAddress := fmt.Sprintf(travelURLTemplate, s.callbackURL, newCustomer.WalletAddress)
	if newCustomer.TravelAddress, err = lnurl.LNURLEncode(travelAddress); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": err.Error()})
		return
	}
	fmt.Printf("\n\nRegister Request:\n%+v\n\n", newCustomer)

	// Save the new Customer struct to the database
	if db := s.db.Create(&newCustomer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not register customer": db.Error})
		return
	}
	c.IndentedJSON(http.StatusCreated, &newCustomer)
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

func (s *server) OriginatorTransfer(c *gin.Context) {
	// Bind the request JSON to a Payload struct
	var err error
	var newPayload Payload
	if err = c.BindJSON(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"!!Could not bind request": err.Error()})
		return
	}

	// Validate the received Payload struct
	if err = validatePayload(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid payload provided": err.Error()})
		return
	}

	// Unmarshal the Payload struct's IVMS101 payload to a
	// trisa.IdentityPayload struct
	ivms101 := &trisa.IdentityPayload{}
	jsonpb := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	if err = jsonpb.Unmarshal([]byte(newPayload.IVMS101), ivms101); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not unmarshal ivms101": err.Error()})
		return
	}

	// Construct the Transfer struct
	newTransfer := Transfer{
		TransferID:     uuid.New(),
		Status:         Pending,
		OriginatorVasp: originatorVasp(ivms101),
		Originator:     originatorName(ivms101),
		Beneficiary:    beneficiaryName(ivms101),
		AssetType:      translateAsset(newPayload.Asset.Slip0044),
		Amount:         newPayload.Amount,
		Created:        time.Now(),
	}

	// Save the new Transfer struct to the database
	if db := s.db.Create(&newTransfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not create transfer": db.Error})
		return
	}

	body := fmt.Sprintf(`{"ivms101": "%s", "asset": %s, "amount": %f, "callback": "%s"}`,
		newPayload.IVMS101,
		newPayload.Asset,
		newPayload.Amount,
		newPayload.Callback)

	var url string
	if url, err = lnurl.LNURLDecode(c.Param("lnurl")); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid LNURL provided": err.Error()})
		return
	}

	var response string
	if response, err = postRequest(body, url); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not create transfer": err})
		return
	}
	fmt.Println(response)
	c.IndentedJSON(http.StatusOK, response)
}

func translateAsset(assetType string) (asset VirtualAsset) {
	switch slip0044 := assetType; slip0044 {
	case "BTC":
		asset = Bitcoin
	case "BTH":
		asset = BitcoinCash
	case "ETH":
		asset = Ethereum
	case "LTC":
		asset = Litecoin
	case "XRP":
		asset = Ripple
	case "XTZ":
		asset = Tezos
	case "EOS":
		asset = EOS
	default:
		asset = UnknownAsset
	}
	return asset
}

// The Transfer endpoint initiates a TRP transfer,
// validaating the transfer payload and saving the
// pending transfer GORM model to the Postgres database.
func (s *server) Transfer(c *gin.Context) {
	// Bind the request JSON to a Payload struct
	var err error
	var newPayload Payload
	if err = c.BindJSON(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	newPayload.IVMS101 = strings.ReplaceAll(newPayload.IVMS101, `*`, `"`)
	newPayload.IVMS101 = strings.ReplaceAll(newPayload.IVMS101, "+", "\n")

	// Validate the received Payload struct
	if err = validatePayload(&newPayload); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid payload provided": err.Error()})
		return
	}

	ivms101 := trisa.IdentityPayload{}
	if err = json.Unmarshal([]byte(newPayload.IVMS101), &ivms101); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid payload provided": err.Error()})
	}
	fmt.Println(&ivms101)

	// Construct the Transfer struct
	newTransfer := Transfer{
		TransferID:     uuid.New(),
		Status:         Pending,
		OriginatorVasp: originatorVasp(&ivms101),
		Originator:     originatorName(&ivms101),
		Beneficiary:    beneficiaryName(&ivms101),
		AssetType:      translateAsset(newPayload.Asset.Slip0044),
		Amount:         newPayload.Amount,
		Created:        time.Now(),
	}

	// Save the new Transfer struct to the database
	if db := s.db.Create(&newTransfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not create transfer": db.Error})
		return
	}

	var data []byte
	if data, err = json.Marshal(newTransfer); err != nil {
		fmt.Printf("Marshal error: %s", err)
	}
	fmt.Printf("\n%s\n\n", data)

	// Respond with approval or rejection

	c.IndentedJSON(http.StatusAccepted,
		&TransferReply{
			Approved: &TransferApproval{
				Address:  "payment address",
				Callback: fmt.Sprintf(confirmationURLTemplate, s.callbackURL, newTransfer.TransferID),
			},
		})

}

func (s *server) ListTransfers(c *gin.Context) {
	var transfers []Transfer
	s.db.Find(&transfers)
	c.IndentedJSON(http.StatusOK, &transfers)
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
	if db := s.db.Where("transfer_id = ?", TransferID).First(&transfer); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not find transfer": db.Error})
	}
	c.IndentedJSON(http.StatusOK, &transfer)
}

func (s *server) OriginatorConfirmation(c *gin.Context) {
	// Bind the request JSON to a TransferReply struct
	var err error
	var reply TransferReply
	if err = c.BindJSON(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	fmt.Println(reply)

	// Validate the received TransferReply struct
	if err = validateReply(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid resolution": err.Error()})
		return
	}

	// Update the transfer status
	var body string
	transferID := c.Param("id")
	transfer := &Transfer{}
	if db := s.db.Model(transfer).Where("transfer_id = ?", transferID); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
	}
	// if transfer.Status != Pending {
	// 	c.IndentedJSON(http.StatusBadRequest, gin.H{"Transfer not in pending state": transfer})
	// 	return
	// }
	if reply.Approved != nil {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
			return
		}
		body = fmt.Sprintf(`{"approved": {"address": "payment address", "callback": "%s"}}`, reply.Approved.Callback)
	} else if reply.Rejected != "" {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Rejected); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not reject transfer": db.Error})
			return
		}
		body = fmt.Sprintln(`{"rejected": "transfer canceled"}`)
	}

	var response string
	if response, err = postRequest(body, reply.Approved.Callback); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not create transfer": err})
		return
	}
	fmt.Println(response)
	c.IndentedJSON(http.StatusOK, response)
}

// The TransferConfirmation endpoint handles and validates
// callbacks that  are exucuted based on the callback url
// provided by requests to the InquiryResolution endpoint.
// These callbacks should provide either asset specific
// identifiers resulting from on-chain transfer executions
// or transfer cancelation comments.
func (s *server) TransferConfirmation(c *gin.Context) {
	// Bind the request JSON to a TransferConfirmation struct
	var err error
	var reply TransferReply
	if err = c.BindJSON(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Could not bind request": err.Error()})
		return
	}

	// Validate the received TransferConfirmation struct
	if err = validateReply(&reply); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Invalid resolution": err.Error()})
		return
	}

	// Update the transfer status
	transferID := c.Param("id")
	transfer := &Transfer{}
	if db := s.db.Model(transfer).Where("transfer_id = ?", transferID); db.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
		return
	}
	// if transfer.Status != Pending {
	// 	c.IndentedJSON(http.StatusBadRequest, gin.H{"Transfer not in pending state": transfer})
	// 	return
	// }
	if reply.Rejected == "" {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Approved); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not approve transfer": db.Error})
			return
		}
		c.IndentedJSON(http.StatusOK, &TransferConfirmation{TxId: transferID})
		return
	} else {
		if db := s.db.Model(&Transfer{}).Where("transfer_id = ?", transferID).Update("Status", Rejected); db.Error != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Could not reject transfer": db.Error})
			return
		}
		c.IndentedJSON(http.StatusOK, &TransferConfirmation{Canceled: reply.Rejected})
		return
	}
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

// Helper function to ensure that the JSON provided to the transfer
// endpoint is valid
func validatePayload(payload *Payload) (err error) {
	if payload.IVMS101 == "" {
		return errors.New("ivms101 payload must be set")
	}

	if payload.Asset.Slip0044 == "" {
		return errors.New("asset must be set")
	}

	if payload.Amount == 0 {
		return errors.New("transfer amount must be set")
	}

	if payload.Callback == "" {
		return errors.New("callback must be set")
	}
	return nil
}

// Helper function to ensure that the JSON provided to the
// inquiryresolution endpoint is valid
func validateReply(reply *TransferReply) error {
	if reply.Approved != nil {
		if reply.Approved.Callback == "" {
			return errors.New("approved replies must have a callback")
		}
	}

	err := errors.New("reply must either be approved or rejected")
	if reply.Approved == nil && reply.Rejected == "" {
		return err
	}
	if reply.Approved != nil && reply.Rejected != "" {
		return err
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

// sends a POST request containing the provided body to the provided
// URL and returns the response
func postRequest(body string, url string) (_ string, err error) {
	var request *http.Request
	byteBody := []byte(body)
	if request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(byteBody)); err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	var response *http.Response
	if response, err = http.DefaultClient.Do(request); err != nil {
		return "", err
	}

	responseBody := &bytes.Buffer{}
	if _, err = io.Copy(responseBody, response.Body); err != nil {
		return "", err
	}
	return responseBody.String(), nil
}
