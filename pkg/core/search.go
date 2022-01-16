package core

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/blugelabs/bluge"
	v1 "github.com/prabhatsharma/zinc/pkg/meta/v1"
	"github.com/prabhatsharma/zinc/pkg/uquery"
)

func (ind *Index) Search(q v1.ZincQuery) (v1.SearchResponse, error) {
	var Hits []v1.Hit

	var searchRequest bluge.SearchRequest

	if q.MaxResults == 0 {
		q.MaxResults = 20
	}

	var err error

	switch q.SearchType {
	case "alldocuments":
		searchRequest, err = uquery.AllDocuments(q)
	case "wildcard":
		searchRequest, err = uquery.WildcardQuery(q)
	case "fuzzy":
		searchRequest, err = uquery.FuzzyQuery(q)
	case "term":
		searchRequest, err = uquery.TermQuery(q)
	case "daterange":
		searchRequest, err = uquery.DateRangeQuery(q)
	case "matchall":
		searchRequest, err = uquery.MatchAllQuery(q)
	case "match":
		searchRequest, err = uquery.MatchQuery(q)
	case "matchphrase":
		searchRequest, err = uquery.MatchPhraseQuery(q)
	case "multiphrase":
		searchRequest, err = uquery.MultiPhraseQuery(q)
	case "prefix":
		searchRequest, err = uquery.PrefixQuery(q)
	case "querystring":
		searchRequest, err = uquery.QueryStringQuery(q)
	}

	if err != nil {
		return v1.SearchResponse{Error: err.Error()}, err
	}

	// sample time range aggregation start
	// timestampAggregation := aggregations.DateRanges(search.Field("@timestamp"))
	// daterange1 := aggregations.NewDateRange(time.Now().Add(-time.Hour*24*30), time.Now())
	// timestampAggregation.AddRange(daterange1)
	// searchRequest.AddAggregation("@timestamp", timestampAggregation)
	// sample time range aggregation end

	writer := ind.Writer

	reader, err := writer.Reader()
	if err != nil {
		log.Printf("error accessing reader: %v", err)
	}

	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Printf("error executing search: %v", err)
	}

	// highlighter := highlight.NewANSIHighlighter()

	// iterationStartTime := time.Now()
	next, err := dmi.Next()
	for err == nil && next != nil {
		var result map[string]interface{}
		var id string
		var timestamp time.Time
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			if field == "_source" {
				json.Unmarshal(value, &result)
				return true
			} else if field == "_id" {
				id = string(value)
				return true
			} else if field == "@timestamp" {
				timestamp, _ = bluge.DecodeDateTime(value)
				return true
			}
			return true
		})
		if err != nil {
			log.Printf("error accessing stored fields: %v", err)
		}

		hit := v1.Hit{
			Index:     ind.Name,
			Type:      ind.Name,
			ID:        id,
			Score:     next.Score,
			Timestamp: timestamp,
			Source:    result,
		}

		next, err = dmi.Next()
		// results = append(results, result)

		Hits = append(Hits, hit)
	}
	if err != nil {
		log.Printf("error iterating results: %v", err)
	}

	// fmt.Println("Got results after data load from disk in: ", time.Since(iterationStartTime))
	resp := v1.SearchResponse{
		// Took: int(time.Since(searchStart).Milliseconds()),
		Took:     int(dmi.Aggregations().Duration().Milliseconds()),
		MaxScore: dmi.Aggregations().Metric("max_score"),
		// Buckets:  dmi.Aggregations().Buckets("@timestamp"),
		Hits: v1.Hits{
			Total: v1.Total{
				Value: int(dmi.Aggregations().Count()),
			},
			Hits: Hits,
		},
	}

	reader.Close()

	return resp, nil
}
