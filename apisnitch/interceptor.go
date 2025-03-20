package apiwhisperer

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
)

const (
	// RequestIDKey is the key used to store the request ID in the context
	RequestIDKey = "api-snitch-reqid"
)

// ApiSnitchPlugin is a middleware function to intercept all HTTP requests
func ApiSnitchPlugin(secret string) func(next echo.HandlerFunc) echo.HandlerFunc {
	// Init cache & reporter
	reporter := NewReporter(secret, NewApiCallCache())
	go func() {
		err := reporter.Start()
		if err != nil {
			slog.Error("Failed to connect to API Snitch server", "error", err)
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Perform actions before the request is processed
			slog.Info("Intercepted request", "path", c.Request().URL.Path)

			// Append a unique identifier to the request
			ctx := context.WithValue(c.Request().Context(), RequestIDKey, uuid.New().String())
			// Set the new context to the request
			c.SetRequest(c.Request().WithContext(ctx))

			// Wrap the response writer
			writer := newResponseWriter(c.Response())
			c.Response().Writer = writer

			recordEchoRequest(reporter, c)
			defer finalizeEchoRequest(reporter, c)

			// Call the next handler
			return next(c)
		}
	}
}

func recordEchoRequest(reporter Reporter, c echo.Context) {
	// get request id from context
	reqID := c.Request().Context().Value(RequestIDKey).(string)
	slog.Info("INIT Request ID", "reqID", reqID)

	// get api path
	apiPath := c.Path()
	slog.Info("API path", "apiPath", apiPath)

	// get api request method
	apiMethod := c.Request().Method
	slog.Info("Request method", "apiMethod", apiMethod)

	// TODO per request or per call?
	// get request host
	requestHost := c.Request().Host
	slog.Info("Request host", "requestHost", requestHost)

	// get request size
	requestBodySize := c.Request().ContentLength
	slog.Info("Request body size", "requestBodySize", requestBodySize)

	// get actual request path
	requestPath := c.Request().URL.Path
	slog.Info("Request path", "requestPath", requestPath)

	// get request headers
	requestHeaders := c.Request().Header
	slog.Info("Request headers", "requestHeaders", requestHeaders)

	// get request query parameters
	requestQueryParams := c.Request().URL.Query()
	slog.Info("Request query parameters", "requestQueryParams", requestQueryParams)

	// get request form values
	requestFormValues := c.Request().Form
	slog.Info("Request form values", "requestFormValues", requestFormValues)

	// get request remote address
	requestRemoteAddress := c.Request().RemoteAddr
	slog.Info("Request remote address", "requestRemoteAddress", requestRemoteAddress)

	// get request user agent
	requestUserAgent := c.Request().UserAgent()
	slog.Info("Request user agent", "requestUserAgent", requestUserAgent)

	// get request cookies
	requestCookies := cookiesFromRequest(c.Request().Cookies())

	// get request referer
	requestReferer := c.Request().Referer()
	slog.Info("Request referer", "requestReferer", requestReferer)

	// get request protocol
	requestProtocol := c.Request().Proto
	slog.Info("Request protocol", "requestProtocol", requestProtocol)

	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		return
	}
	requestBody := string(bodyBytes)

	reporter.CreateApiCall(reqID, apiPath, apiMethod, Request{
		Host:          requestHost,
		Protocol:      requestProtocol,
		Path:          requestPath,
		BodySize:      requestBodySize,
		Headers:       requestHeaders,
		QueryParams:   requestQueryParams,
		Body:          requestBody,
		FormValues:    requestFormValues,
		RemoteAddress: requestRemoteAddress,
		UserAgent:     requestUserAgent,
		Cookies:       requestCookies,
		Referer:       requestReferer,
	})
}

func cookiesFromRequest(cookies []*http.Cookie) map[string]string {
	cookieMap := make(map[string]string)
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}
	return cookieMap
}

func finalizeEchoRequest(reporter Reporter, c echo.Context) {
	// if this defer approach does not work, try using Response().After(func)
	reqID := c.Request().Context().Value(RequestIDKey).(string)
	slog.Info("FINALIZE Request ID", "reqID", reqID)

	responseBodySize := c.Response().Size
	slog.Info("Response body size", "responseBodySize", responseBodySize)

	responseBody := c.Response().Writer.(*responseWriter).body.String()
	slog.Debug("Response body", "responseBody", responseBody)

	response := Response{
		Status:   c.Response().Status,
		BodySize: responseBodySize,
		Headers:  c.Response().Header(),
		Body:     responseBody,
	}
	reporter.FinalizeApiCall(reqID, response)

	slog.Info("Echo request finalized", "reqID", reqID)
}

type responseWriter struct {
	echo.Response
	body *bytes.Buffer
}

func newResponseWriter(w *echo.Response) *responseWriter {
	return &responseWriter{
		Response: *w,
		body:     new(bytes.Buffer),
	}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.Response.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}
