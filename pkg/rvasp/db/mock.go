package db

import (
	"database/sql"
	"encoding/json"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBMock creates a new mocked database object which uses an underlying SQL mock.
// The mock can be safely passed as a real database object to high-level functions
// which interact with gorm.DB objects to enable testing.
func NewDBMock(vasp string) (d *DB, mock sqlmock.Sqlmock, err error) {
	d = &DB{}

	if d.db, mock, err = gormMock(); err != nil {
		return nil, nil, err
	}

	var person []byte
	if person, err = json.Marshal(&ivms101.Person{}); err != nil {
		return nil, nil, err
	}

	// Fetch mocked VASP info from the "database"
	query := regexp.QuoteMeta(`SELECT * FROM "vasps" WHERE name = $1`)
	mock.ExpectQuery(query).WithArgs(vasp).WillReturnRows(mock.NewRows([]string{"id", "name", "ivms101"}).AddRow(42, vasp, string(person)))

	if err = d.db.Where("name = ?", vasp).First(&d.vasp).Error; err != nil {
		return nil, nil, err
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		return nil, nil, err
	}

	return d, mock, nil
}

func gormMock() (d *gorm.DB, mock sqlmock.Sqlmock, err error) {
	var mockDB *sql.DB

	if mockDB, mock, err = sqlmock.New(); err != nil {
		return nil, nil, err
	}

	if d, err = gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	})); err != nil {
		return nil, nil, err
	}

	return d, mock, nil
}
