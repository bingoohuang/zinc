package core

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"github.com/jeremywohl/flatten"
	"github.com/prabhatsharma/zinc/pkg/zutils"
	"github.com/rs/zerolog/log"

	"github.com/blugelabs/bluge"
)

// BuildBlugeDocFromJSON returns the bluge document for the json document. It also updates the mapping for the fields if not found.
// If no mappings are found, it creates te mapping for all the encountered fields. If mapping for some fields is found but not for others
// then it creates the mapping for the missing fields.
func (ind *Index) BuildBlugeDocFromJSON(docID string, doc *map[string]interface{}) (*bluge.Document, error) {
	// Pick the index mapping from the cache if it already exists
	indexMapping := ind.CachedMapping
	if indexMapping == nil {
		indexMapping = make(map[string]string)
	}

	flatDoc, _ := flatten.Flatten(*doc, "", flatten.DotStyle)

	// Create a new bluge document
	bdoc := bluge.NewDocument(docID)

	indexMappingNeedsUpdate := false

	// Iterate through each field and add it to the bluge document
	for key, value := range flatDoc {
		if _, ok := indexMapping[key]; !ok {
			// Assign auto inferred type for the new key
			// fmt.Println("Missing Existing Key in index: ", key)

			// Use reflection to find the type of the value.
			// Bluge requires the field type to be specified.

			if value != nil { // value could be just {} in the json data or e.g. "rules": null or "creationTimestamp": null, etc.
				v := reflect.ValueOf(value)

				// try to find the type of the value and use it to define default mapping
				switch v.Type().String() {
				case "string":
					indexMapping[key] = "text"
				case "float64":
					indexMapping[key] = "numeric"
				case "bool":
					indexMapping[key] = "keyword"
				case "time.Time":
					indexMapping[key] = "time"
				}

				indexMappingNeedsUpdate = true
			}
		}

		if value != nil {
			switch indexMapping[key] {
			case "text": // found using existing index mapping
				f := bluge.NewTextField(key, value.(string)).SearchTermPositions()
				bdoc.AddField(f)
			case "numeric": // found using existing index mapping
				f := bluge.NewNumericField(key, value.(float64))
				bdoc.AddField(f)
			case "keyword": // found using existing index mapping
				f := bluge.NewKeywordField(key, strconv.FormatBool(value.(bool)))
				bdoc.AddField(f)
			case "time": // found using existing index mapping
				f := bluge.NewDateTimeField(key, value.(time.Time))
				bdoc.AddField(f)
			}
		}
	}

	if indexMappingNeedsUpdate {
		ind.SetMapping(indexMapping)
	}

	docByteVal, _ := json.Marshal(*doc)
	bdoc.AddField(bluge.NewDateTimeField("@timestamp", time.Now()).StoreValue())
	bdoc.AddField(bluge.NewStoredOnlyField("_source", docByteVal))
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil)) // Add _all field that can be used for search

	return bdoc, nil
}

// SetMapping Saves the mapping of the index to _index_mapping index
// index: Name of the index for which the mapping needs to be saved
// iMap: a map of the fields at specify name and type of the field. e.g. movietitle: string
func (ind *Index) SetMapping(iMap map[string]string) error {
	// Create a new bluge document
	bdoc := bluge.NewDocument(ind.Name)

	for k, v := range iMap {
		bdoc.AddField(bluge.NewTextField(k, v).StoreValue())
	}

	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))

	// update on the disk
	systemIndex := ZincSystemIndexList["_index_mapping"].Writer
	err := systemIndex.Update(bdoc.ID(), bdoc)
	if err != nil {
		log.Printf("error updating document: %v", err)
		return err
	}

	// update in the cache
	ind.CachedMapping = iMap

	return nil
}

// GetStoredMapping returns the mappings of all the indexes from _index_mapping system index
func (ind *Index) GetStoredMapping() (map[string]string, error) {
	dataPath := zutils.GetEnv("ZINC_DATA_DIR", "./data")

	systemPath := dataPath + "/_index_mapping"

	config := bluge.DefaultConfig(systemPath)

	reader, err := bluge.OpenReader(config)
	if err != nil {
		return nil, nil // probably no system index available
		// log.Fatalf("GetIndexMapping: unable to open reader: %v", err)
	}

	// search for the index mapping _index_mapping index
	query := bluge.NewTermQuery(ind.Name).SetField("_id")
	searchRequest := bluge.NewTopNSearch(1, query) // Should get just 1 result at max

	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Log().Msg("error executing search: " + err.Error())
	}

	next, err := dmi.Next()
	if err != nil {
		return nil, err
	}

	if next != nil {
		result := make(map[string]string)

		err = next.VisitStoredFields(func(field string, value []byte) bool {
			result[field] = string(value)
			return true
		})
		if err != nil {
			return nil, err
		}

		return result, nil

	}

	reader.Close()

	return nil, nil
}
