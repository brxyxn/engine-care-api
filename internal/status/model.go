package status

type Response struct {
	Message   string `json:"message"`
	SvrStatus string `json:"server_status"`
	DBStatus  string `json:"database_status"`
}
