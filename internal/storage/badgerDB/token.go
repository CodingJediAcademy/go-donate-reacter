package badgerDB

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/oauth2"
)

func (s *Storage) SaveToken(token *oauth2.Token) error {
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

func (s *Storage) GetToken() (*oauth2.Token, error) {
	var token oauth2.Token
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("token"))
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

	return &token, err
}
