package api

var (
	GET   = "GET"
	POST  = "POST"
	PATCH = "PATCH"
	PUT   = "PUT"
	DEL   = "DELETE"
)

type Response[T any] struct {
	Code   int            `json:"code"`
	Status string         `json:"status"`
	Result T              `json:"result,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Stack   interface{} `json:"stack,omitempty"`
	Message string      `json:"message"`
}
