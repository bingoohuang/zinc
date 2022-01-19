package core

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/prabhatsharma/zinc/pkg/zutil"
)

const (
	SystemIndexUsers   string = "_users"
	SystemIndexMapping string = "_index_mapping"
)

var systemIndexList = []string{SystemIndexUsers, SystemIndexMapping}

func LoadZincSystemIndexes() (map[string]*Index, error) {
	log.Print("Loading system indexes...")

	IndexList := make(map[string]*Index)
	for _, systemIndex := range systemIndexList {
		idx, err := NewIndex(systemIndex, Disk)
		if err != nil {
			log.Printf("Error loading system index %s error : %v", systemIndex, err.Error())
			return nil, err
		}
		IndexList[systemIndex] = idx
		IndexList[systemIndex].IndexType = "system"
		log.Printf("Index %s loaded", systemIndex)
	}

	return IndexList, nil
}

func LoadZincIndexesFromDisk() (map[string]*Index, error) {
	log.Print("Loading indexes... from disk")

	indexList := make(map[string]*Index)
	files, err := os.ReadDir(zutil.GetDataDir())
	if err != nil {
		log.Fatalf("Error reading data directory: %v", err)
	}

	for _, f := range files {
		iName := f.Name()
		if isSystemIndex := zutil.SliceContains(systemIndexList, iName); isSystemIndex {
			continue
		}

		if idx, err := NewIndex(iName, Disk); err != nil {
			log.Printf("Error loading index: %s, error: %v", iName, err) // inform and move in to next index
		} else {
			indexList[iName] = idx
			indexList[iName].IndexType = "user"
			log.Print("Index loaded: " + iName)
		}
	}

	return indexList, nil
}

func LoadZincIndexesFromS3() (map[string]*Index, error) {
	bucket := zutil.GetS3Bucket()
	if bucket == "" {
		return nil, nil
	}

	log.Print("Loading indexes from s3...")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Print("Error loading AWS config: ", err)
	}
	client := s3.NewFromConfig(cfg)
	IndexList := make(map[string]*Index)

	delimiter := "/"

	ctx := context.Background()
	params := s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Delimiter: &delimiter,
	}

	val, err := client.ListObjectsV2(ctx, &params)
	if err != nil {
		log.Print("failed to list indexes in s3: ", err.Error())
		return nil, err
	}

	for _, obj := range val.CommonPrefixes {
		iName := (*obj.Prefix)[0 : len(*obj.Prefix)-1]
		idx, err := NewIndex(iName, S3)

		if err != nil {
			log.Print("failed to load index "+iName+" in s3: ", err.Error())
		} else {
			IndexList[iName] = idx
			IndexList[iName].IndexType = "user"
			IndexList[iName].StorageType = S3
			log.Print("Index loaded: " + iName)
		}
	}

	return IndexList, nil
}
