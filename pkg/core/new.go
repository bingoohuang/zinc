package core

import (
	"github.com/blugelabs/bluge"
	"github.com/prabhatsharma/zinc/pkg/dir"
	"github.com/prabhatsharma/zinc/pkg/zutil"
)

type StorageType string

const (
	Disk StorageType = "disk"
	S3   StorageType = "s3"
)

// NewIndex creates an instance of a physical zinc index that can be used to store and retrieve data.
func NewIndex(name string, storageType StorageType) (*Index, error) {
	config := func(storageType StorageType) bluge.Config {
		if storageType == S3 {
			return dir.GetS3Config(zutil.GetS3Bucket(), name)
		} else { // Default storage type is disk
			return bluge.DefaultConfig(zutil.GetDataDir() + "/" + name)
		}
	}(storageType)

	writer, err := bluge.OpenWriter(config)
	if err != nil {
		return nil, err
	}

	index := &Index{
		Name:        name,
		Writer:      writer,
		StorageType: storageType,
	}

	mapping, err := index.GetStoredMapping()
	if err != nil {
		return nil, err
	}

	index.CachedMapping = mapping
	return index, nil
}
