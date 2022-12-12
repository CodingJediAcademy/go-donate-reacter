package badgerDB

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"go-donate-reacter/internal/services/donationalerts/response"
)

func (s *Storage) SaveProfile(profile *response.Profile) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(profile); err != nil {
			return err
		}

		if err := txn.Set([]byte("profile"), buf.Bytes()); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *Storage) GetProfile() (*response.Profile, error) {
	var profile response.Profile
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("profile"))
		if err != nil {
			return err
		}

		gobValue, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(gobValue)
		dec := gob.NewDecoder(buf)

		if err := dec.Decode(&profile); err != nil {
			return err
		}

		return nil
	})

	return &profile, err
}
