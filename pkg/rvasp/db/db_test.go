package db_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
)

const (
	FIXTURES_PATH = "../fixtures"
	NUM_TABLES    = 5
	NUM_INDICES   = 11
)

// Expect a query which does no row updates (e.g. CREATE TABLE, DROP TABLE, etc.)
func expectExec(mock sqlmock.Sqlmock, query string) {
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 0))
}

// Expect a number of CREATE queries
func expectCreate(mock sqlmock.Sqlmock, numCreated int) {
	for i := 0; i < numCreated; i++ {
		expectExec(mock, "CREATE")
	}
}

// Expect a number of DROP TABLE queries
func expectDropTable(mock sqlmock.Sqlmock, numDropped int) {
	for i := 0; i < numDropped; i++ {
		expectExec(mock, "DROP TABLE IF EXISTS")
	}
}

// Expect a query which inserts a number of rows into a table
func expectInsert(mock sqlmock.Sqlmock, table string, numInserted int) {
	if numInserted > 1 {
		mock.ExpectBegin()
	}

	rows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < numInserted; i++ {
		rows = rows.AddRow(i + 1)
	}

	mock.ExpectQuery(fmt.Sprintf(`INSERT INTO "%s"`, table)).WillReturnRows(rows)

	if numInserted > 1 {
		mock.ExpectCommit()
	}
}

// dbTestSuite tests interactions with the high-level database functions and asserts
// that the expected SQL queries are executed.
type dbTestSuite struct {
	suite.Suite
	db   *db.DB
	mock sqlmock.Sqlmock
}

func TestDB(t *testing.T) {
	suite.Run(t, new(dbTestSuite))
}

func (s *dbTestSuite) BeforeTest(suiteName, testName string) {
	var err error
	s.db, s.mock, err = db.NewDBMock("alice")
	require.NoError(s.T(), err)
}

func (s *dbTestSuite) AfterTest() {
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *dbTestSuite) TestMigrateDB() {
	require := s.Require()
	gdb := s.db.GetDB()

	// We can't rely on the order in which the migration operations are executed, so
	// this asserts that the number of operations is correct.
	expectCreate(s.mock, NUM_TABLES+NUM_INDICES)
	require.NoError(db.MigrateDB(gdb))
}

func (s *dbTestSuite) TestResetDB() {
	// Expecting the tables to be dropped
	expectDropTable(s.mock, NUM_TABLES)

	// Expecting the tables and indices to be created
	expectCreate(s.mock, NUM_TABLES+NUM_INDICES)

	// Expecting table inserts
	expectInsert(s.mock, "vasps", 3)
	expectInsert(s.mock, "wallets", 12)
	expectInsert(s.mock, "accounts", 12)

	// Reset the database
	require.NoError(s.T(), db.ResetDB(s.db.GetDB(), FIXTURES_PATH))
}

func (s *dbTestSuite) TestLookupAccount() {
	require := s.Require()
	email := "mary@alicevasp.us"
	id := s.db.GetVASP().ID

	// Account lookups should be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE (vasp_id = $1 AND email = $2 OR wallet_address = $3)`)
	s.mock.ExpectQuery(query).WithArgs(id, email, email).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var account db.Account
	tx := s.db.LookupAccount(email).First(&account)
	require.NoError(tx.Error)
	require.Equal(id, account.ID)
}

func (s *dbTestSuite) TestLookupAnyAccount() {
	require := s.Require()
	email := "mary@alicevasp.us"
	id := s.db.GetVASP().ID

	// LookupAnyAccount should not be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE (email = $1 OR wallet_address = $2)`)
	s.mock.ExpectQuery(query).WithArgs(email, email).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var account db.Account
	tx := s.db.LookupAnyAccount(email).First(&account)
	require.NoError(tx.Error)
	require.Equal(id, account.ID)
}

func (s *dbTestSuite) TestLookupBeneficiary() {
	require := s.Require()
	beneficiary := "mary@alicevasp.us"
	id := s.db.GetVASP().ID

	// Beneficiary lookups should be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "wallets" WHERE (vasp_id = $1 AND email = $2 OR address = $3)`)
	s.mock.ExpectQuery(query).WithArgs(id, beneficiary, beneficiary).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var wallet db.Wallet
	tx := s.db.LookupBeneficiary(beneficiary).First(&wallet)
	require.NoError(tx.Error)
	require.Equal(id, wallet.ID)
}

func (s *dbTestSuite) TestLookupAnyBeneficiary() {
	require := s.Require()
	beneficiary := "mary@alicevasp.us"
	id := s.db.GetVASP().ID

	// LookupAnyBeneficiary lookups should not be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "wallets" WHERE (email = $1 OR address = $2)`)
	s.mock.ExpectQuery(query).WithArgs(beneficiary, beneficiary).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var wallet db.Wallet
	tx := s.db.LookupAnyBeneficiary(beneficiary).First(&wallet)
	require.NoError(tx.Error)
	require.Equal(id, wallet.ID)
}

func (s *dbTestSuite) TestLookupIdentity() {
	require := s.Require()
	address := "18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh"
	id := s.db.GetVASP().ID

	// Identity lookups should be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "identities" WHERE vasp_id = $1 AND wallet_address = $2`)
	s.mock.ExpectQuery(query).WithArgs(id, address).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var identity db.Identity
	tx := s.db.LookupIdentity(address).First(&identity)
	require.NoError(tx.Error)
	require.Equal(id, identity.ID)
}

func (s *dbTestSuite) TestLookupPending() {
	require := s.Require()
	id := s.db.GetVASP().ID

	// Transaction lookups should be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE vasp_id = $1 AND state in ($2, $3, $4)`)
	s.mock.ExpectQuery(query).WithArgs(id, pb.TransactionState_PENDING_SENT, pb.TransactionState_PENDING_RECEIVED, pb.TransactionState_PENDING_ACKNOWLEDGED).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var transaction db.Transaction
	tx := s.db.LookupPending().First(&transaction)
	require.NoError(tx.Error)
	require.Equal(id, transaction.ID)
}

func (s *dbTestSuite) TestLookupWallet() {
	require := s.Require()
	address := "18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh"
	id := s.db.GetVASP().ID

	// Wallet lookups should be limited to the configured VASP
	query := regexp.QuoteMeta(`SELECT * FROM "wallets" WHERE vasp_id = $1 AND address = $2`)
	s.mock.ExpectQuery(query).WithArgs(id, address).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

	var wallet db.Wallet
	tx := s.db.LookupWallet(address).First(&wallet)
	require.NoError(tx.Error)
	require.Equal(id, wallet.ID)
}

func (s *dbTestSuite) TestTransactions() {
	require := s.Require()
	email := "mary@alicevasp.us"
	accountID := uint(47)
	vaspID := s.db.GetVASP().ID
	transactionIDs := []uint{1, 2, 3}

	// Fetch mocked Account from the database
	query := regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE (vasp_id = $1 AND email = $2 OR wallet_address = $3)`)
	s.mock.ExpectQuery(query).WithArgs(vaspID, email, email).WillReturnRows(s.mock.NewRows([]string{"id"}).AddRow(accountID))

	var account db.Account
	tx := s.db.LookupAccount(email).First(&account)
	require.NoError(tx.Error)
	require.Equal(accountID, account.ID)
	require.NoError(s.mock.ExpectationsWereMet())

	// Fetch mock transactions for the mocked Account
	query = regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE vasp_id = $1 AND account_id = $2`)
	rows := sqlmock.NewRows([]string{"id"})
	for _, id := range transactionIDs {
		rows = rows.AddRow(id)
	}
	s.mock.ExpectQuery(query).WithArgs(vaspID, account.ID).WillReturnRows(rows)

	transactions, err := account.Transactions(s.db)
	require.NoError(err)
	require.Len(transactions, len(transactionIDs))
}
