package db

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/trisacrypto/testnet/pkg/rvasp/jsonpb"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DB is a wrapper around a gorm.DB instance that restricts query results to a single
// VASP.
type DB struct {
	db   *gorm.DB
	vasp *VASP
}

func NewDB(dsn string, vasp string) (d *DB, err error) {
	d = &DB{}
	if d.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}); err != nil {
		return nil, err
	}

	if err = MigrateDB(d.db); err != nil {
		return nil, err
	}

	if err = d.db.Where("name = ?", vasp).First(&d.vasp).Error; err != nil {
		return nil, fmt.Errorf("could not fetch VASP info from database: %s", err)
	}

	if vasp != d.vasp.Name {
		return nil, fmt.Errorf("expected name %q but have database name %q", vasp, d.vasp.Name)
	}

	return d, nil
}

func (d *DB) Query() *gorm.DB {
	return d.db.Where("vasp_id = ?", d.vasp.ID)
}

func (d *DB) Create(value interface{}) *gorm.DB {
	return d.db.Create(value)
}

func (d *DB) Save(value interface{}) *gorm.DB {
	return d.db.Save(value)
}

// LookupAccount by email address or wallet address.
func (d *DB) LookupAccount(account string) *gorm.DB {
	return d.Query().Where("email = ?", account).Or("wallet_address = ?", account)
}

// LookupBeneficiary by email address or wallet address.
func (d *DB) LookupBeneficiary(beneficiary string) *gorm.DB {
	return d.Query().Preload("Provider").Where("email = ?", beneficiary).Or("address = ?", beneficiary)
}

// VASP is a record of known partner VASPs and caches TRISA protocol information. This
// table also contains IVMS101 data that identifies the VASP (but only for the local
// VASP - we assume that VASPs do not have IVMS101 data on each other and have to use
// the directory service for that).
// TODO: modify VASP ID to a GUID
type VASP struct {
	gorm.Model
	Name      string     `gorm:"uniqueIndex;size:255;not null"`
	LegalName *string    `gorm:"column:legal_name;null"`
	URL       *string    `gorm:"null"`
	Country   *string    `gorm:"null"`
	Endpoint  *string    `gorm:"null"`
	PubKey    *string    `gorm:"null"`
	NotAfter  *time.Time `gorm:"null"`
	IVMS101   string     `gorm:"column:ivms101"`
}

// TableName explicitly defines the name of the table for the model
func (VASP) TableName() string {
	return "vasps"
}

// Wallet is a mapping of wallet IDs to VASPs to determine where to send transactions.
// Provider lookups can happen by wallet address or by email.
type Wallet struct {
	gorm.Model
	Address    string `gorm:"uniqueIndex"`
	Email      string `gorm:"uniqueIndex"`
	ProviderID uint   `gorm:"not null"`
	Provider   VASP   `gorm:"foreignKey:ProviderID"`
	VaspID     uint   `gorm:"not null"`
	Vasp       VASP   `gorm:"foreignKey:VaspID"`
}

// TableName explicitly defines the name of the table for the model
func (Wallet) TableName() string {
	return "wallets"
}

// Account contains details about the transactions that are served by the local VASP.
// It also contains the IVMS 101 data for KYC verification, in this table it is just
// stored as a JSON string rather than breaking it down to the field level. Only
// customers of the VASP have accounts.
type Account struct {
	gorm.Model
	Name          string          `gorm:"not null"`
	Email         string          `gorm:"uniqueIndex;not null"`
	WalletAddress string          `gorm:"uniqueIndex;not null;column:wallet_address"`
	Wallet        Wallet          `gorm:"foreignKey:WalletAddress;references:Address"`
	Balance       decimal.Decimal `gorm:"type:numeric(15,2);default:0.0"`
	Completed     uint64          `gorm:"not null;default:0"`
	Pending       uint64          `gorm:"not null;default:0"`
	IVMS101       string          `gorm:"column:ivms101;not null"`
	VaspID        uint            `gorm:"not null"`
	Vasp          VASP            `gorm:"foreignKey:VaspID"`
}

// TableName explicitly defines the name of the table for the model
func (Account) TableName() string {
	return "accounts"
}

// Transaction holds exchange information to send money from one account to another. It
// also contains the decrypted identity payload that was sent as part of the TRISA
// protocol and the envelope ID that uniquely identifies the message chain.
type Transaction struct {
	gorm.Model
	Envelope      string          `gorm:"not null"`
	AccountID     uint            `gorm:"not null"`
	Account       Account         `gorm:"foreignKey:AccountID"`
	OriginatorID  uint            `gorm:"column:originator_id;not null"`
	Originator    Identity        `gorm:"foreignKey:OriginatorID"`
	BeneficiaryID uint            `gorm:"column:beneficiary_id;not null"`
	Beneficiary   Identity        `gorm:"foreignKey:BeneficiaryID"`
	Amount        decimal.Decimal `gorm:"type:numeric(15,2)"`
	Debit         bool            `gorm:"not null"`
	Completed     bool            `gorm:"not null;default:false"`
	Timestamp     time.Time       `gorm:"not null"`
	Identity      string          `gorm:"not null"`
	VaspID        uint            `gorm:"not null"`
	Vasp          VASP            `gorm:"foreignKey:VaspID"`
}

// TableName explicitly defines the name of the table for the model
func (Transaction) TableName() string {
	return "transactions"
}

// Identity holds raw data for an originator or a beneficiary that was sent as
// part of the transaction process. This should not be stored in the wallet since the
// wallet is a representation of the local VASPs knowledge about customers and bercause
// the identity information could change between transactions. This intermediate table
// is designed to more closely mimic data storage as part of a blockchain transaction.
type Identity struct {
	gorm.Model
	WalletAddress string `gorm:"not null;column:wallet_address"`
	Email         string `gorm:"uniqueIndex"`
	Provider      string `gorm:"not null"`
	VaspID        uint   `gorm:"not null"`
	Vasp          VASP   `gorm:"foreignKey:VaspID"`
}

// TableName explicitly defines the name of the table for the model
func (Identity) TableName() string {
	return "identities"
}

// BalanceFloat converts the balance decmial into an exact two precision float32 for
// use with the protocol buffers.
func (a Account) BalanceFloat() float32 {
	bal, _ := a.Balance.Truncate(2).Float64()
	return float32(bal)
}

// Transactions returns an ordered list of transactions associated with the account
// ordered by the timestamp of the transaction, listing any pending transactions at the
// top. This function may also support pagination and limiting functions, which is why
// we're using it rather than having a direct relationship on the model.
func (a Account) Transactions(db *DB) (records []Transaction, err error) {
	if err = db.Query().Preload(clause.Associations).Where("account_id = ?", a.ID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// AmountFloat converts the amount decmial into an exact two precision float32 for
// use with the protocol buffers.
func (t Transaction) AmountFloat() float32 {
	bal, _ := t.Amount.Truncate(2).Float64()
	return float32(bal)
}

// Proto converts the transaction into a protocol buffer transaction
func (t Transaction) Proto() *pb.Transaction {
	return &pb.Transaction{
		Originator: &pb.Account{
			WalletAddress: t.Originator.WalletAddress,
			Email:         t.Originator.Email,
			Provider:      t.Originator.Provider,
		},
		Beneficiary: &pb.Account{
			WalletAddress: t.Beneficiary.WalletAddress,
			Email:         t.Beneficiary.Email,
			Provider:      t.Beneficiary.Provider,
		},
		Amount:    t.AmountFloat(),
		Timestamp: t.Timestamp.Format(time.RFC3339),
		Envelope:  t.Envelope,
		Identity:  t.Identity,
	}
}

// LoadIdentity returns the ivms101.Person for the VASP.
func (v VASP) LoadIdentity() (person *ivms101.Person, err error) {
	if v.IVMS101 == "" {
		return nil, fmt.Errorf("vasp record %d does not have IVMS101 data", v.ID)
	}

	person = new(ivms101.Person)
	if err = jsonpb.UnmarshalString(v.IVMS101, person); err != nil {
		return nil, fmt.Errorf("could not unmarshal identity: %s", err)
	}
	return person, nil
}

// LoadIdentity returns the ivms101.Person for the Account.
func (a Account) LoadIdentity() (person *ivms101.Person, err error) {
	if a.IVMS101 == "" {
		return nil, fmt.Errorf("account record %d does not have IVMS101 data", a.ID)
	}

	person = new(ivms101.Person)
	if err = jsonpb.UnmarshalString(a.IVMS101, person); err != nil {
		return nil, fmt.Errorf("could not unmarshal identity: %s", err)
	}
	return person, nil
}

// MigrateDB the schema based on the models defined above.
func MigrateDB(gdb *gorm.DB) (err error) {
	// Migrate models
	if err = gdb.AutoMigrate(&VASP{}, &Wallet{}, &Account{}, &Transaction{}, &Identity{}); err != nil {
		return err
	}

	return nil
}

// ResetDB resets the database using the JSON fixtures.
func ResetDB(gdb *gorm.DB, fixturesPath string) (err error) {
	var (
		vasps    []VASP
		wallets  []Wallet
		accounts []Account
	)

	// Load the VASP fixtures
	if vasps, err = LoadVASPs(fixturesPath); err != nil {
		return err
	}

	// Load the wallet and account fixtures
	if wallets, accounts, err = LoadWallets(fixturesPath); err != nil {
		return err
	}

	// Reset the database
	if err = gdb.Migrator().DropTable(&VASP{}, &Wallet{}, &Account{}, &Transaction{}, &Identity{}); err != nil {
		return err
	}

	// Migration to create the tables
	if err = MigrateDB(gdb); err != nil {
		return err
	}

	// Insert the VASP fixtures into the database
	if err = gdb.Create(&vasps).Error; err != nil {
		return err
	}

	// Insert the wallet fixtures into the database
	if err = gdb.Create(&wallets).Error; err != nil {
		return err
	}

	// Insert the account fixtures into the database
	if err = gdb.Create(&accounts).Error; err != nil {
		return err
	}

	return nil
}
