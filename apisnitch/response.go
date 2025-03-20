package apisnitch

type Response struct {
	Status   int                 `json:"status"`
	BodySize int64               `json:"bodySize"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body"`
}
