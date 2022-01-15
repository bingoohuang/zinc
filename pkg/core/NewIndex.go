package core

import (
	"github.com/blugelabs/bluge"
	"github.com/prabhatsharma/zinc/pkg/directory"
	"github.com/prabhatsharma/zinc/pkg/zutils"
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
			s := zutils.GetEnv("S3_BUCKET", "")
			return directory.GetS3Config(s, name)
		} else { // Default storage type is disk
			s := zutils.GetEnv("ZINC_DATA_DIR", "./data")
			return bluge.DefaultConfig(s + "/" + name)
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

func IndexExists(index string) (bool, StorageType) {
	if _, ok := ZincIndexList[index]; ok {
		return true, ZincIndexList[index].StorageType
	}

	return false, ""
}
