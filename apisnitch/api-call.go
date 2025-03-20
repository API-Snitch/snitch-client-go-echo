package apiwhisperer

type ApiCall struct {
	ApiPath       string   `json:"path"`
	ApiMethod     string   `json:"method"`
	RequestID     string   `json:"reqId"`
	Request       Request  `json:"req"`
	Response      Response `json:"res"`
	Timestamp     int64    `json:"ts"`
	DurationMicro int64    `json:"dur"`
	Finalized     bool     `json:"finalized"`
}

func NewApiCall(apiPath string, apiMethod string, requestID string, timestamp int64) ApiCall {
	return ApiCall{
		RequestID: requestID,
		ApiPath:   apiPath,
		ApiMethod: apiMethod,
		Timestamp: timestamp,
	}
}
