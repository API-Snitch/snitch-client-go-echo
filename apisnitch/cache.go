package apiwhisperer

import (
	"log/slog"
	"time"
)

type apiCallCache struct {
	cacheMap map[string]ApiCall
}

type ApiCallCache interface {
	CreateApiCall(requestID string, apiPath string, apiMethod string, request Request)
	FinalizeApiCall(requestID string, response Response) *ApiCall
	GetFinalized() []ApiCall
	DeleteApiCall(requestID string) error
}

func NewApiCallCache() ApiCallCache {
	return &apiCallCache{
		cacheMap: make(map[string]ApiCall),
	}
}

func (c *apiCallCache) CreateApiCall(requestID string, apiPath string, apiMethod string, request Request) {
	apiCall := NewApiCall(apiPath, apiMethod, requestID, time.Now().UnixMilli())
	apiCall.Request = request
	c.cacheMap[requestID] = apiCall
}

func (c *apiCallCache) FinalizeApiCall(requestID string, response Response) *ApiCall {
	call, found := c.cacheMap[requestID]
	if !found {
		slog.Warn("ApiCall not found", "requestID", requestID)
		return nil
	}
	call.Response = response
	call.DurationMicro = time.Now().UnixMicro() - call.Timestamp*1000
	call.Finalized = true
	c.cacheMap[requestID] = call
	slog.Info("ApiCall finalized", "requestID", requestID, "size", len(c.cacheMap))
	return &call
}

func (c *apiCallCache) GetFinalized() []ApiCall {
	finalized := make([]ApiCall, 0)
	for _, call := range c.cacheMap {
		if call.Finalized {
			finalized = append(finalized, call)
		}
	}
	return finalized
}

func (c *apiCallCache) DeleteApiCall(requestID string) error {
	_, found := c.cacheMap[requestID]
	if !found {
		return nil
	}
	delete(c.cacheMap, requestID)
	return nil
}
