package definitions

type Email struct {
	Date    string `json:"date"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailList struct {
	Index  string  `json:"index"`
	Emails []Email `json:"records"`
}

type Query struct {
	Search     string   `json:"search"`
	SortFields []string `json:"sort_fields"`
	Source     string   `json:"source"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
}

type SearchQuery struct {
	Term      string `json:"term"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type Search struct {
	SearchType string      `json:"search_type"`
	Query      SearchQuery `json:"query"`
	SortFields []string    `json:"sort_fields"`
	From       int         `json:"from"`
	MaxResults int         `json:"max_results"`
	Source     []string    `json:"_source"`
}

type shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type hits struct {
	Total struct {
		Value int `json:"value"`
	} `json:"total"`
	Hits []struct {
		ID     string `json:"_id"`
		Source Email  `json:"_source"`
	} `json:"hits"`
}

type SearchResponse struct {
	Took     int    `json:"took"`
	TimedOut bool   `json:"timed_out"`
	Shards   shards `json:"_shards"`
	Hits     hits   `json:"hits"`
}
