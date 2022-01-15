package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/blugelabs/bluge"
	"github.com/gin-gonic/gin"
	"github.com/prabhatsharma/zinc/pkg/core"
)

func UpdateDoc(c *gin.Context) {
	indexName := c.Param("target")
	queryId := c.Param("id") // ID for the document to be updated provided in URL path

	var doc map[string]interface{}
	c.BindJSON(&doc)

	docID := ""
	// If id field is present then use it, else create a new UUID and use it
	if id, ok := doc["_id"]; ok {
		docID = id.(string)
	} else if queryId != "" {
		docID = queryId
	}

	mintedID := docID == ""
	if mintedID {
		docID = uuid.New().String() // Generate a new ID if ID was not provided
	}

	// If the index does not exist, then create it
	if exists, _ := core.IndexExists(indexName); !exists {
		newIndex, err := core.NewIndex(indexName, core.Disk) // Create a new index with disk storage as default
		if err != nil {
			log.Print(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		core.ZincIndexList[indexName] = newIndex // Load the index in memory
	}

	index := core.ZincIndexList[indexName]
	if err := index.UpdateDoc(docID, &doc, mintedID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"id": docID})
	}
}

func DeleteDoc(c *gin.Context) {
	indexName := c.Param("target")
	queryId := c.Param("id")

	if indexExists, _ := core.IndexExists(indexName); !indexExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index not exist"})
		return
	}

	bdoc := bluge.NewDocument(queryId)
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))
	docIndexWriter := core.ZincIndexList[indexName].Writer
	if err := docIndexWriter.Delete(bdoc.ID()); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Deleted", "index": indexName, "id": queryId})
	}
}
