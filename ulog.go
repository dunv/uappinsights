package uappinsights

import (
	"bytes"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
)

type LogWriter struct {
	client          appinsights.TelemetryClient
	appName         string
	fieldMapping    [3][2]int
	severityMapping map[string]contracts.SeverityLevel
}

func NewLogWriter(client appinsights.TelemetryClient, appName string) *LogWriter {
	return &LogWriter{
		client:  client,
		appName: appName,
		fieldMapping: [3][2]int{
			{26, 32},
			{34},
		},
		severityMapping: map[string]contracts.SeverityLevel{
			"TRACE": contracts.Verbose,
			"DEBUG": contracts.Verbose,
			"INFO":  contracts.Information,
			"WARN":  contracts.Warning,
			"ERROR": contracts.Error,
			"FATAL": contracts.Critical,
		},
	}
}

func (w *LogWriter) SetFieldMapping(fieldMapping [3][2]int) {
	w.fieldMapping = fieldMapping
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	log_level := bytes.TrimSpace(p[w.fieldMapping[0][0]:w.fieldMapping[0][1]])
	message := bytes.TrimSpace(p[w.fieldMapping[1][0]:])
	d := appinsights.NewTraceTelemetry(string(message), w.severityMapping[string(log_level)])
	d.Properties[CUSTOM_TAG] = w.appName
	w.client.Track(d)
	return len(p), nil
}
