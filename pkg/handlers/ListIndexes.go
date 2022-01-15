package handlers

import (
	"net/http"

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
