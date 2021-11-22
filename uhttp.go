package uappinsights

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/dunv/uhttp"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

// AppInsightsMiddleware logs requests to appInsights
func AppInsightsMiddleware(client appinsights.TelemetryClient) func(next http.HandlerFunc) http.HandlerFunc {
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
			client.TrackRequest(r.Method, r.URL.EscapedPath(), duration, strconv.Itoa(airw.statusCode))
		}
	}
}

type appInsightsResponseWriter struct {
	w           http.ResponseWriter
	wroteHeader bool
	statusCode  int
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

func ForwardPanics(client appinsights.TelemetryClient) func(r *http.Request, err error) {
	return func(r *http.Request, err error) {
		client.TrackException(err)
	}
}
