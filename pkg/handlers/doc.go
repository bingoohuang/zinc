package handlers

import (
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/blugelabs/bluge"
	"github.com/gin-gonic/gin"
	"github.com/prabhatsharma/zinc/pkg/core"
)

func UpdateDoc(c *gin.Context) {
	index, err := core.GetIndex(c.Param("target"))
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	var doc map[string]interface{}
	c.BindJSON(&doc)

	docID, mintedID := parseDocID(doc, c.Param("id"))
	if err := index.UpdateDoc(docID, &doc, mintedID); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"id": docID})
	}
}

// parseDocID parse id field is present then use it, else create a new UUID and use it.
func parseDocID(doc map[string]interface{}, queryId string) (string, bool) {
	docID := ""
	if id, ok := doc["_id"]; ok {
		docID = id.(string)
	} else if queryId != "" {
		docID = queryId
	}

	if docID != "" {
		return docID, false
	}

	return uuid.New().String(), true
}

func DeleteDoc(c *gin.Context) {
	indexName := c.Param("target")
	queryId := c.Param("id")
	index, ok := core.FindIndex(indexName)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index not exist"})
		return
	}

	bdoc := bluge.NewDocument(queryId)
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))
	if err := index.Writer.Delete(bdoc.ID()); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Deleted", "index": indexName, "id": queryId})
	}
}
