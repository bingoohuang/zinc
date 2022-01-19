package handlers

import (
	"bufio"
	"encoding/json"
	"io"
	"log"

	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/prabhatsharma/zinc/pkg/zutil"

	"github.com/blugelabs/bluge/index"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prabhatsharma/zinc/pkg/core"
)

func BulkHandler(c *gin.Context) {
	result, err := BulkHandlerWorker(c.Param("target"), c.Request.Body)
	if err != nil {
		c.JSON(200, gin.H{"message": err})
		return
	}

	c.JSON(200, gin.H{
		"message":     "bulk data inserted",
		"updateCount": result.UpdateCount,
		"insertCount": result.InsertCount,
	})
}

type BulkResult struct {
	UpdateCount int
	InsertCount int
}

const _index = "_index"

func BulkHandlerWorker(target string, body io.ReadCloser) (*BulkResult, error) {
	// Prepare to read the entire raw text of the body
	scanner := bufio.NewScanner(body)

	// Set 1 MB max per line. docs at - https://pkg.go.dev/bufio#pkg-constants
	// This is the max size of a line in a file that we will process
	const maxCapacityPerLine = 1024 * 1024
	buf := make([]byte, maxCapacityPerLine)
	scanner.Buffer(buf, maxCapacityPerLine)

	nextLineIsData := false
	lastLineMetaData := make(map[string]interface{})

	batch := make(map[string]*index.Batch)
	var indexesInThisBatch []string
	var bulkResult BulkResult

	for scanner.Scan() { // Read each line
		var doc map[string]interface{}
		err := json.Unmarshal(scanner.Bytes(), &doc) // Read each line as JSON and store it in doc
		if err != nil {
			log.Print(err)
		}

		// This will process the data line in the request. Each data line is preceded by a metadata line.
		// Docs at https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
		if nextLineIsData {
			nextLineIsData = false
			id := ""
			mintedID := false

			if val, ok := lastLineMetaData["id"]; ok {
				id = val.(string)
			} else {
				id = uuid.New().String()
				mintedID = true
			}

			indexName := lastLineMetaData[_index].(string)
			// Since this is a bulk request, we need to check if we already created a new batch for this index. We need to create 1 batch per index.
			if !zutil.SliceContains(indexesInThisBatch, indexName) { // Add the list of indexes to the batch if it's not already there
				indexesInThisBatch = append(indexesInThisBatch, indexName)
				batch[indexName] = index.NewBatch()
			}

			idx, err := core.GetIndex(indexName)
			if err != nil {
				return nil, err
			}

			bdoc, err := idx.BuildBlugeDocFromJSON(id, &doc)
			// Add the document to the batch. We will persist the batch to the index
			// when we have processed all documents in the request
			if !mintedID {
				batch[indexName].Update(bdoc.ID(), bdoc)
				bulkResult.UpdateCount++
			} else {
				batch[indexName].Insert(bdoc)
				bulkResult.InsertCount++
			}

		} else { // This branch will process the metadata line in the request. Each metadata line is preceded by a data line.
			for k, v := range doc {
				switch k {
				case "index", "create", "update", "delete":
					vm, _ := v.(map[string]interface{})
					lastLineMetaData["operation"] = k
					lastLineMetaData["_id"] = vm["_id"]
					// if index is specified in metadata then it overtakes the index in the query path
					lastLineMetaData[_index] = ss.Or(vm[_index].(string), target)
					nextLineIsData = k != "delete"
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, n := range indexesInThisBatch {
		writer := core.ZincIndexList[n].Writer

		// Persist the batch to the index
		if err := writer.Batch(batch[n]); err != nil {
			log.Print("Error updating batch: ", err.Error())
			return nil, err
		}
	}

	return &bulkResult, nil
}
