package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/blugelabs/bluge"
	v1 "github.com/prabhatsharma/zinc/pkg/meta/v1"
	"github.com/prabhatsharma/zinc/pkg/zutil"

	"github.com/gin-gonic/gin"
	"github.com/prabhatsharma/zinc/pkg/core"
)

func ListIndexes(c *gin.Context) {
	indexListMap := make(map[string]*SimpleIndex)
	for name, value := range core.ZincIndexList {
		indexListMap[name] = &SimpleIndex{
			Name:          name,
			CachedMapping: value.CachedMapping,
		}
	}
	c.JSON(http.StatusOK, indexListMap)
}

type SimpleIndex struct {
	Name          string            `json:"name"`
	CachedMapping map[string]string `json:"mapping"`
}

// SearchIndex searches the index for the given http request from end user
func SearchIndex(c *gin.Context) {
	indexName := c.Param("target")
	if indexExists, _ := core.IndexExists(indexName); !indexExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index '" + indexName + "' does not exist"})
		return
	}

	var query v1.ZincQuery
	c.BindJSON(&query)

	index := core.ZincIndexList[indexName]
	res, errS := index.Search(query)
	if errS != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errS.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func CreateIndex(c *gin.Context) {
	var newIndex core.Index
	c.BindJSON(&newIndex)

	cIndex, err := core.NewIndex(newIndex.Name, newIndex.StorageType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	core.ZincIndexList[newIndex.Name] = cIndex

	c.JSON(http.StatusOK, gin.H{
		"result":       "Index: " + newIndex.Name + " created",
		"storage_type": newIndex.StorageType,
	})
}

// DeleteIndex deletes a zinc index and its associated data. Be careful using thus as you ca't undo this action.
func DeleteIndex(c *gin.Context) {
	indexName := c.Param("indexName")

	// 0. Check if index exists and Get the index storage type - disk, s3 or memory
	indexExists, indexStorageType := core.IndexExists(indexName)
	if !indexExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index " + indexName + "does not exist"})
		return
	}

	// 1. Close the index writer
	core.ZincIndexList[indexName].Writer.Close()

	// 2. Delete from the cache
	delete(core.ZincIndexList, indexName)

	// 3. Physically delete the index
	switch indexStorageType {
	case core.Disk:
		if err := os.RemoveAll(zutil.GetDataDir() + "/" + indexName); err != nil {
			log.Print("failed to delete index: ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case core.S3:
		if err := deleteFilesForIndexFromS3(indexName); err != nil {
			log.Print("failed to delete index: ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	// 4. Delete the index mapping
	bdoc := bluge.NewDocument(indexName)
	if err := core.ZincSystemIndexList["_index_mapping"].Writer.Delete(bdoc.ID()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Deleted",
			"index":   indexName,
			"storage": indexStorageType,
		})
	}
}

func deleteFilesForIndexFromS3(indexName string) error {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Print("Error loading AWS config: ", err)
		return err
	}
	client := s3.NewFromConfig(cfg)

	s3Bucket := zutil.GetS3Bucket()
	ctx := context.Background()

	// List Objects in the bucket at prefix
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: &s3Bucket,
		Prefix: &indexName,
	}
	listObjectsOutput, err := client.ListObjectsV2(ctx, listObjectsInput)
	if err != nil {
		log.Print("failed to list objects: ", err.Error())
		return err
	}

	var fileList []types.ObjectIdentifier

	for _, object := range listObjectsOutput.Contents {
		fileList = append(fileList, types.ObjectIdentifier{
			Key: object.Key,
		})
		fmt.Println("Deleting: ", *object.Key)
	}

	doi := &s3.DeleteObjectsInput{
		Bucket: &s3Bucket,
		Delete: &types.Delete{
			Objects: fileList,
		},
	}
	_, err = client.DeleteObjects(ctx, doi)

	if err != nil {
		log.Print("failed to delete index: ", err.Error())
		return err
	}

	return nil
}
