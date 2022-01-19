package core

import (
	"github.com/blugelabs/bluge"
)

var (
	// ZincIndexList Nothing to handle in the error. If you can't load indexes then everything is broken.
	ZincIndexList       map[string]*Index
	ZincSystemIndexList map[string]*Index
)

func FindIndex(index string) (*Index, bool) {
	if v, ok := ZincIndexList[index]; ok {
		return v, true
	}

	return nil, false
}

// GetIndex gets or creates a new index by the index name.
func GetIndex(indexName string) (*Index, error) {
	v, ok := ZincIndexList[indexName]
	if ok {
		return v, nil
	}

	idx, err := NewIndex(indexName, Disk)
	if err != nil {
		return nil, err
	}

	ZincIndexList[indexName] = idx // Load the index in memory
	return idx, nil
}

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
