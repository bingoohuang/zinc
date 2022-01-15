package uquery

import (
	"fmt"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis/analyzer"
	qs "github.com/blugelabs/query_string"
	v1 "github.com/prabhatsharma/zinc/pkg/meta/v1"
)

func WildcardQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	// requestedIndex := startup.INDEX_WRITER_LIST[indexName]
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	wildcardQuery := bluge.NewWildcardQuery(iQuery.Query.Term).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(wildcardQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func TermQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	termQuery := bluge.NewTermQuery(iQuery.Query.Term).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(termQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func QueryStringQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	options := qs.DefaultOptions()
	options.WithDefaultAnalyzer(analyzer.NewStandardAnalyzer())
	userQuery, err := qs.ParseQueryString(iQuery.Query.Term, options)
	if err != nil {
		return nil, fmt.Errorf("error parsing query string '%s': %v", iQuery.Query.Term, err)
	}

	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")
	finalQuery := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(userQuery)

	// sortFields := []string{"-@timestamp"} // adding a - (minus) before the field name will sort the field in descending order

	searchRequest := buildRequest(iQuery, finalQuery)

	return searchRequest, nil
}

func PrefixQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	prefixQuery := bluge.NewPrefixQuery(iQuery.Query.Term).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(prefixQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func MultiPhraseQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	multiPhraseQuery := bluge.NewMultiPhraseQuery(iQuery.Query.Terms).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(multiPhraseQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func MatchQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")
	dateQuery.SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	matchQuery := bluge.NewMatchQuery(iQuery.Query.Term).SetField(field).SetAnalyzer(analyzer.NewStandardAnalyzer())
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(matchQuery)

	searchRequest := buildRequest(iQuery, query)
	return searchRequest, nil
}

func MatchPhraseQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	matchPhraseQuery := bluge.NewMatchPhraseQuery(iQuery.Query.Term).SetField(field).SetAnalyzer(analyzer.NewStandardAnalyzer())
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(matchPhraseQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func MatchAllQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	fuzzyQuery := bluge.NewFuzzyQuery(iQuery.Query.Term).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(fuzzyQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func FuzzyQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")

	var field string
	if iQuery.Query.Field != "" {
		field = iQuery.Query.Field
	} else {
		field = "_all"
	}

	fuzzyQuery := bluge.NewFuzzyQuery(iQuery.Query.Term).SetField(field)
	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(fuzzyQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func DateRangeQuery(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")
	query := bluge.NewBooleanQuery().AddMust(dateQuery)

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

func AllDocuments(iQuery v1.ZincQuery) (bluge.SearchRequest, error) {
	dateQuery := bluge.NewDateRangeQuery(iQuery.Query.StartTime, iQuery.Query.EndTime).SetField("@timestamp")
	allQuery := bluge.NewMatchAllQuery()

	query := bluge.NewBooleanQuery().AddMust(dateQuery).AddMust(allQuery)

	// iQuery.MaxResults = 20

	searchRequest := buildRequest(iQuery, query)

	return searchRequest, nil
}

// buildRequest combines the ZincQuery with the bluge Query to create a SearchRequest
func buildRequest(iQuery v1.ZincQuery, query bluge.Query) bluge.SearchRequest {
	return bluge.NewTopNSearch(iQuery.MaxResults, query).
		SetFrom(iQuery.From).
		SortBy(iQuery.SortFields).
		WithStandardAggregations()
}
