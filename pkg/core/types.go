package core

import (
	"github.com/blugelabs/bluge"
)

// ZincIndexList Nothing to handle in the error. If you can't load indexes then everything is broken.
var (
	ZincIndexList, _       = LoadZincIndexesFromDisk()
	ZincSystemIndexList, _ = LoadZincSystemIndexes()
)

func init() {
	s3List, _ := LoadZincIndexesFromS3()
	for k, v := range s3List {
		ZincIndexList[k] = v
	}
}

type Index struct {
	Name          string            `json:"name"`
	Writer        *bluge.Writer     `json:"blugeWriter"`
	CachedMapping map[string]string `json:"mapping"`
	IndexType     string            `json:"index_type"`   // "system" or "user"
	StorageType   StorageType       `json:"storage_type"` // disk, memory, s3
}
