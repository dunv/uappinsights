package uappinsights

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dunv/uhttp"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

// AppInsightsMiddleware logs requests to appInsights
func AppInsightsMiddleware(client appinsights.TelemetryClient, appName string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == uhttp.NO_LOG_MAGIC_URL_FORCE_CACHE {
				next.ServeHTTP(w, r)
				return
			}
			airw := appInsightsResponseWriter{w: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(&airw, r)
			duration := time.Since(start)
			d := appinsights.NewRequestTelemetry(r.Method, r.URL.EscapedPath(), duration, strconv.Itoa(airw.statusCode))

			// Add custom tag (for discrimination of multiple services within on instance of appinsights)
			d.Properties[CUSTOM_TAG] = appName

			// Add all query-params to the request to make it easier to figure out what went wrong
			for k, v := range r.URL.Query() {
				d.Properties[fmt.Sprintf("query__%s", k)] = strings.Join(v, ",")
			}

			// Add first couple of characters, in case the request failed
			if airw.statusCode != http.StatusOK {
				d.Properties["response__error"] = fmt.Sprintf("%s [...]", airw.errorResponse)
			}

			client.Track(d)
		}
	}
}

type appInsightsResponseWriter struct {
	w             http.ResponseWriter
	wroteHeader   bool
	statusCode    int
	errorResponse string
}

// Delegate Header() to underlying responseWriter
func (w *appInsightsResponseWriter) Header() http.Header {
	return w.w.Header()
}

// Delegate Write() to underlying responseWriter
func (w *appInsightsResponseWriter) Write(data []byte) (int, error) {
	// the default implementation in net/http/server.go (line 1577 in go 1.17.2) writes the response-header as
	// soon as write is called, if there are no headers written yet
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.statusCode != http.StatusOK {
		if len(data) > 200 {
			w.errorResponse = string(data[:200])
		} else {
			w.errorResponse = string(data)
		}
	}
	return w.w.Write(data)
}

// Delegate WriteHeader() to underlying responseWriter AND save code
func (w *appInsightsResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.w.WriteHeader(code)
	}
}

// Delegate Hijack() to underlying responseWriter
func (w *appInsightsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func ForwardPanics(client appinsights.TelemetryClient, appName string) func(r *http.Request, err error) {
	return func(r *http.Request, err error) {
		d := appinsights.NewExceptionTelemetry(err)
		d.Properties[CUSTOM_TAG] = appName
		client.Track(d)
	}
}
