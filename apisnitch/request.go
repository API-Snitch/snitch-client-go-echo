package apisnitch

type Request struct {
	Host          string              `json:"host"`
	Path          string              `json:"path"`
	Method        string              `json:"method"`
	Protocol      string              `json:"protocol"`
	QueryParams   map[string][]string `json:"queryParams"`
	BodySize      int64               `json:"bodySize"`
	Body          string              `json:"body"`
	Headers       map[string][]string `json:"headers"`
	FormValues    map[string][]string `json:"formValues"`
	RemoteAddress string              `json:"remoteAddress"`
	UserAgent     string              `json:"userAgent"`
	Cookies       map[string]string   `json:"cookies"`
	Referer       string              `json:"referer"`
}
