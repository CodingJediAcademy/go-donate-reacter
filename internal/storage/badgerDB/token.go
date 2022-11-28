package badgerDB

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"go-donate-reacter/internal/services/donationalerts/response"
)

func (s *Storage) SaveToken(token response.Token) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(token); err != nil {
			return err
		}

		if err := txn.Set([]byte("token"), buf.Bytes()); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *Storage) GetToken() (response.Token, error) {
	var token response.Token
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("answer"))
		if err != nil {
			return err
		}

		gobValue, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(gobValue)
		dec := gob.NewDecoder(buf)

		if err := dec.Decode(&token); err != nil {
			return err
		}

		return nil
	})

	return token, err
}
