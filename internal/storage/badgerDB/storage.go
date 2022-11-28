package badgerDB

import (
	"github.com/dgraph-io/badger/v3"
	"log"
	"os"
	"path/filepath"
)

type Storage struct {
	db *badger.DB
}

func NewStorage() Storage {
	return Storage{}
}

func (s *Storage) Init() {
	filePath := s.dataDirInit()

	opt := badger.DefaultOptions(filePath)
	opt.BlockCacheSize = 100 << 20

	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}

	s.db = db
}

func (s *Storage) Close() {
	if err := s.db.Close(); err != nil {
		log.Fatal(err)
	}
}

func (s *Storage) dataDirInit() string {
	userDirectory, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	dirPath := filepath.Join(userDirectory, "donate-reacter", "data")
	if err := os.MkdirAll(dirPath, 0666); err != nil {
		log.Fatal(err)
	}

	return dirPath
}
