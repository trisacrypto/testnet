package db_test

import (
	"io"
	"io/ioutil"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Copies the SQLite database in the testdata directory to a temporary file location and
// opens the temporary database file (so that modifications to the database will not be
// preserved across tests). Also returns a cleanup function that will close the database
// connection and remove temporary files, the cleanup function must be called when the
// test finishes using the testdb fixture.
func testdb() (db *gorm.DB, cleanup func() error, err error) {
	// Copy the testdb from testdata into a temporary directory
	var tmpdb string
	if tmpdb, err = copydb(); err != nil {
		return nil, nil, err
	}

	if db, err = gorm.Open(sqlite.Open(tmpdb), &gorm.Config{}); err != nil {
		// Cleanup temporary database before returning the error
		os.Remove(tmpdb)
		return nil, nil, err
	}

	cleanup = func() error {
		if err = os.Remove(tmpdb); err != nil {
			return err
		}
		return nil
	}

	return db, cleanup, nil
}

func copydb() (path string, err error) {
	var src, dst *os.File
	if src, err = os.Open("testdata/test.db"); err != nil {
		return "", err
	}
	defer src.Close()

	if dst, err = ioutil.TempFile("", "rvasp_test_*.db"); err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return dst.Name(), nil
}

/*
func TestMigrateDB(t *testing.T) {
	// Create the test database fixture
	gdb, cleanup, err := testdb()
	require.NoError(t, err)
	defer cleanup()

	// Migrate the test database
	err = db.MigrateDB(gdb)
	require.NoError(t, err)

	// Ensure the correct number of rows are still in the DB
	var count int64
	db.Table("vasps").Count(&count)
	require.Equal(t, int64(3), count)

	db.Table("wallets").Count(&count)
	require.Equal(t, int64(9), count)

	db.Table("accounts").Count(&count)
	require.Equal(t, int64(3), count)

	db.Table("transactions").Count(&count)
	require.Equal(t, int64(0), count)

	db.Table("identities").Count(&count)
	require.Equal(t, int64(0), count)
}*/

/*
func TestAccountHelpers(t *testing.T) {
	// Create the test database fixture
	gdb, cleanup, err := testdb()
	require.NoError(t, err)
	defer cleanup()

	// Migrate the test database
	err = MigrateDB(gdb)
	require.NoError(t, err)

	// Test lookup account by email address or wallet
	var aca db.Account
	tx := LookupAccount(db, "mary@alicevasp.us").First(&aca)
	require.NoError(t, tx.Error)
	require.Equal(t, uint(1), aca.ID)

	var acb Account
	tx = LookupAccount(db, "14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v").First(&acb)
	require.NoError(t, tx.Error)
	require.Equal(t, uint(3), acb.ID)
}*/

/*
func TestWalletHelpers(t *testing.T) {
	// Create the test database fixture
	db, cleanup, err := testdb()
	require.NoError(t, err)
	defer cleanup()

	// Migrate the test database
	err = MigrateDB(db)
	require.NoError(t, err)

	// Test lookup wallet by email address or wallet address
	var bna Wallet
	tx := LookupBeneficiary(db, "george@bobvasp.co.uk").First(&bna)
	require.NoError(t, tx.Error)
	require.Equal(t, uint(2), bna.ID)

	var bnb Wallet
	tx = LookupBeneficiary(db, "182kF4mb5SW4KGEvBSbyXTpDWy8rK1Dpu").First(&bnb)
	require.NoError(t, tx.Error)
	require.Equal(t, uint(9), bnb.ID)
}
*/
