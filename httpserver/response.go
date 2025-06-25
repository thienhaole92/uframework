package httpserver

type Response struct {
	RequestID  string      `json:"requestId,omitempty"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Limit     int64 `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"totalPage"`
}
