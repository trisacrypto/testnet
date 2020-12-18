package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
)

// OpenLevelDB directory Store at the specified path. This is the default storage provider.
func OpenLevelDB(uri string) (Store, error) {
	dsn, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("could not parse leveldb uri: %s", err)
	}

	db, err := leveldb.OpenFile(dsn.Path, nil)
	if err != nil {
		return nil, err
	}

	store := &ldbStore{db: db}
	if err = store.sync(); err != nil {
		return nil, err
	}

	return store, nil
}

// Errors that may occur during LevelDB operations
var (
	ErrCorruptedSequence = errors.New("primary key sequence is invalid")
	ErrCorruptedIndex    = errors.New("search indices are invalid")
	ErrIncompleteRecord  = errors.New("vasp record is missing required fields")
	ErrEntityNotFound    = errors.New("entity not found")
	ErrDuplicateEntity   = errors.New("entity unique constraints violated")
)

// keys and prefixes for leveldb buckets and indices
var (
	keyAutoSequence = []byte("pks")
	keyNameIndex    = []byte("names")
	keyCountryIndex = []byte("countries")
	preVASPS        = []byte("vasps")
)

// Implements Store for some basic LevelDB operations and simple protocol buffer storage.
type ldbStore struct {
	sync.RWMutex
	db        *leveldb.DB
	sequence  uint64         // autoincrement sequence for ID values
	names     uniqueIndex    // case insensitive name index
	countries containerIndex // lookup vasps in a specific country
}

// Close the database, allowing no further interactions. This method also synchronizes
// the indices to ensure that they are saved between sessions.
func (s *ldbStore) Close() error {
	defer s.db.Close()
	if err := s.sync(); err != nil {
		return err
	}
	return nil
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

// Create a VASP into the directory. This method requires the VASP to have a unique
// name and ignores any ID fields that are set on the VASP, instead assigning new IDs.
func (s *ldbStore) Create(v pb.VASP) (id string, err error) {
	// Create the name to check the uniqueness constraint
	name := strings.TrimSpace(strings.ToLower(v.CommonName))
	if name == "" {
		printJSON(v)
		return "", ErrIncompleteRecord
	}

	// Update management timestamps
	v.LastUpdated = time.Now().Format(time.RFC3339)
	if v.FirstListed == "" {
		v.FirstListed = v.LastUpdated
	}

	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	// Check the uniqueness constraint
	if _, ok := s.names[name]; ok {
		printJSON(v)
		return "", ErrDuplicateEntity
	}

	var data []byte
	key := s.vaspKey(v.Id)
	if data, err = proto.Marshal(&v); err != nil {
		return "", err
	}

	if err = s.db.Put(key, data, nil); err != nil {
		return "", err
	}

	// Update indices after successful insert
	s.names[name] = v.Id
	s.countries.add(v.Id, v.Entity.CountryOfRegistration)
	return v.Id, nil
}

// Retrieve a VASP record by id; returns an error if the record does not exist.
func (s *ldbStore) Retrieve(id string) (v pb.VASP, err error) {
	var val []byte
	key := s.vaspKey(id)
	if val, err = s.db.Get(key, nil); err != nil {
		if err == leveldb.ErrNotFound {
			return v, ErrEntityNotFound
		}
		return v, err
	}

	if err = proto.Unmarshal(val, &v); err != nil {
		return v, err
	}

	return v, nil
}

// Update the VASP entry by the VASP ID (required). This method simply overwrites the
// entire VASP record and does not update individual fields.
func (s *ldbStore) Update(v pb.VASP) (err error) {
	if v.Id == "" {
		return ErrIncompleteRecord
	}

	// Retrieve the original record to ensure that the indices are updated properly
	key := s.vaspKey(v.Id)
	o, err := s.Retrieve(v.Id)
	if err != nil {
		return err
	}

	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	var val []byte
	if val, err = proto.Marshal(&v); err != nil {
		return err
	}

	// Insert the new record
	if err = s.db.Put(key, val, nil); err != nil {
		return err
	}

	// Update indices if necessary
	if v.CommonName != o.CommonName {
		delete(s.names, o.CommonName)
		s.names[v.CommonName] = v.Id
	}

	if v.Entity.CountryOfRegistration != o.Entity.CountryOfRegistration {
		s.countries.rm(v.Id, o.Entity.CountryOfRegistration)
		s.countries.add(v.Id, v.Entity.CountryOfRegistration)
	}

	return nil
}

// Destroy a record, removing it completely from the database and indices.
func (s *ldbStore) Destroy(id string) (err error) {
	key := s.vaspKey(id)

	// Lookup the record in order to remove data from indices
	record, err := s.Retrieve(id)
	if err != nil {
		if err == ErrEntityNotFound {
			return nil
		}
		return err
	}

	// LevelDB will not return an error if the entity does not exist
	if err = s.db.Delete(key, nil); err != nil {
		return err
	}

	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	// Remove the records from the indices
	delete(s.names, record.CommonName)
	s.countries.rm(record.Id, record.Entity.CountryOfRegistration)
	return nil
}

// Search uses the names and countries index to find VASPS that match the specified
// query. This is a very simple search and is not intended for robust usage. To find a
// VASP by name, a case insensitive search is performed if the query exists in
// any of the VASP entity names. Alternatively a list of names can be given or a country
// or list of countries for case-insensitive exact matches.
func (s *ldbStore) Search(query map[string]interface{}) (vasps []pb.VASP, err error) {
	// A set of records that match the query and need to be fetched
	records := make(map[string]struct{})

	s.RLock()
	// Lookup by name
	names, ok := parseQuery("name", query)
	if ok {
		log.WithField("name", names).Debug("search name query")
		for _, name := range names {
			if id := s.names[name]; id != "" {
				records[id] = struct{}{}
			}
		}
	}

	// Lookup by country
	countries, ok := parseQuery("country", query)
	if ok {
		for _, country := range countries {
			for _, id := range s.countries[country] {
				records[id] = struct{}{}
			}
		}
	}
	s.RUnlock()

	// Perform the lookup of records if there are any
	if len(records) > 0 {
		vasps = make([]pb.VASP, 0, len(records))
		for id := range records {
			var vasp pb.VASP
			if vasp, err = s.Retrieve(id); err != nil {
				if err == ErrEntityNotFound {
					continue
				}
				return nil, err
			}
			vasps = append(vasps, vasp)
		}
	}

	return vasps, nil
}

// creates a []byte key from the vasp id using a prefix to act as a leveldb bucket
func (s *ldbStore) vaspKey(id string) (key []byte) {
	buf := []byte(id)
	key = make([]byte, 0, len(preVASPS)+len(buf))
	key = append(key, preVASPS...)
	key = append(key, buf...)
	return key
}

// Helper indices for quick lookups and cheap constraints
type uniqueIndex map[string]string
type containerIndex map[string][]string

// sync all indices with the underlying database
func (s *ldbStore) sync() (err error) {
	if err = s.seqsync(); err != nil {
		return err
	}

	if err = s.syncnames(); err != nil {
		return err
	}

	if err = s.synccountries(); err != nil {
		return err
	}

	return nil
}

// sync the autoincrement sequence with the leveldb auto sequence key
func (s *ldbStore) seqsync() (err error) {
	var pk uint64
	val, err := s.db.Get(keyAutoSequence, nil)
	if err != nil {
		// If the auto sequence key is not found, simply leave pk to 0
		if err != leveldb.ErrNotFound {
			return err
		}
	} else {
		var n int
		if pk, n = binary.Uvarint(val); n <= 0 {
			log.WithField("n", n).Error("could not parse primary key sequence value")
			return ErrCorruptedSequence
		}
	}

	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	// Local is behind database state, set and return
	if s.sequence <= pk {
		s.sequence = pk
		return nil
	}

	//  Update the database with the local state
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, s.sequence)
	if err = s.db.Put(keyAutoSequence, buf, nil); err != nil {
		log.WithError(err).Error("could not put primary key sequence value")
		return ErrCorruptedSequence
	}

	return nil
}

// sync the names index with the leveldb names key
func (s *ldbStore) syncnames() (err error) {
	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	if s.names == nil {
		// fetch the names from the database
		val, err := s.db.Get(keyNameIndex, nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				s.names = make(uniqueIndex)
				return nil
			}
			return err
		}

		if err = json.Unmarshal(val, &s.names); err != nil {
			log.WithError(err).Error("could not unmarshal names index")
			return ErrCorruptedIndex
		}

	}

	// Put the current names back to the database
	val, err := json.Marshal(s.names)
	if err != nil {
		log.WithError(err).Error("could not marshal names index")
		return ErrCorruptedIndex
	}

	if err = s.db.Put(keyNameIndex, val, nil); err != nil {
		log.WithError(err).Error("could not put names index")
		return ErrCorruptedIndex
	}
	return nil
}

// sync the countries index with the leveldb countries key
func (s *ldbStore) synccountries() (err error) {
	// Critical section (optimizing for safety rather than speed)
	s.Lock()
	defer s.Unlock()

	if s.countries == nil {
		// fetch the countries from the database
		val, err := s.db.Get(keyCountryIndex, nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				s.countries = make(containerIndex)
				return nil
			}
			return err
		}

		if err = json.Unmarshal(val, &s.countries); err != nil {
			log.WithError(err).Error("could not unmarshall country index")
			return ErrCorruptedIndex
		}
	}

	// Put the current countries back to the database
	val, err := json.Marshal(s.countries)
	if err != nil {
		log.WithError(err).Error("could not marshal country index")
		return ErrCorruptedIndex
	}

	if err = s.db.Put(keyCountryIndex, val, nil); err != nil {
		log.WithError(err).Error("could not put country index")
		return ErrCorruptedIndex
	}

	return nil
}

func (c containerIndex) add(id string, country string) {
	if country == "" {
		return
	}

	// make country search case insensitive
	country = strings.ToLower(country)

	arr, ok := c[country]
	if !ok {
		arr = make([]string, 0, 10)
		arr = append(arr, id)
		c[country] = arr
		return
	}

	i := sort.Search(len(arr), func(i int) bool { return arr[i] >= id })
	if i < len(arr) && arr[i] == id {
		// value is already in the array
		return
	}

	arr = append(arr, "")
	copy(arr[i+1:], arr[i:])
	arr[i] = id

	c[country] = arr
}

func (c containerIndex) rm(id string, country string) {
	// make country search case insensitive
	country = strings.ToLower(country)

	arr, ok := c[country]
	if !ok {
		return
	}

	i := sort.Search(len(arr), func(i int) bool { return arr[i] >= id })
	if i < len(arr) && arr[i] == id {
		copy(arr[i:], arr[i+1:])
		arr[len(arr)-1] = ""
		arr = arr[:len(arr)-1]
		c[country] = arr
	}
}

// A helper function to fetch a list of values from a query
func parseQuery(key string, query map[string]interface{}) ([]string, bool) {
	val, ok := query[key]
	if !ok {
		return nil, false
	}

	if vals, ok := val.([]string); ok {
		for i := range vals {
			vals[i] = strings.ToLower(strings.TrimSpace(vals[i]))
		}
		return vals, true
	}

	if vals, ok := val.(string); ok {
		vals = strings.ToLower(strings.TrimSpace(vals))
		return []string{vals}, true
	}

	return nil, false
}
