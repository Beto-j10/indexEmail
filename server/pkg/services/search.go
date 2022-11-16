package services

import (
	def "server/pkg/definitions"
)

func (i *indexService) SearchMail(query *def.Query) (*def.SearchResponse, error) {

	search := &def.Search{
		SearchType: "match",
		Query: def.SearchQuery{
			Term:      query.Search,
			StartTime: "2021-01-01T00:00:00.000Z",
			EndTime:   "2040-01-01T00:00:00.000Z",
		},
		MaxResults: query.PageSize,
		From:       (query.Page - 1) * query.PageSize,
	}

	response, err := i.storage.SearchMail(search)
	if err != nil {
		return nil, err
	}

	return response, nil
}
