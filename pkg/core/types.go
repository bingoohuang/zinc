package core

import (
	"github.com/blugelabs/bluge"
)

var (
	// ZincIndexList Nothing to handle in the error. If you can't load indexes then everything is broken.
	ZincIndexList       map[string]*Index
	ZincSystemIndexList map[string]*Index
)

// Init initializes the zinc.
func Init() {
	ZincIndexList, _ = LoadZincIndexesFromDisk()
	ZincSystemIndexList, _ = LoadZincSystemIndexes()

	s3List, _ := LoadZincIndexesFromS3()
	for k, v := range s3List {
		ZincIndexList[k] = v
	}
}

type Index struct {
	Name          string                `json:"name"`
	Writer        *bluge.Writer         `json:"-"`
	CachedMapping map[string]string     `json:"mapping"`
	IndexType     string                `json:"index_type"` // "system" or "user"
	StorageType   `json:"storage_type"` // disk, memory, s3
}
