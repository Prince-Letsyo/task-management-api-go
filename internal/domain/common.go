package domain

const (
	DefaultPageLimit = 25
	MaxPageLimit     = 100
)

type Filters struct {
	Limit  int32 `json:"limit" query:"limit"`
	Offset int32 `json:"offset" query:"offset"`
	Total  int64 `json:"total"`
}

type Page struct {
	Filters
	Metadata map[string]interface{} `json:"metadata"`
}
