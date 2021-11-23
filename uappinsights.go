package uappinsights

import (
	"errors"
	"fmt"
	"strings"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

const CUSTOM_TAG = "applicationName"

func ParseConnectionString(connectionString string) (*appinsights.TelemetryConfiguration, error) {
	parts := strings.Split(connectionString, ";")

	if len(parts) != 2 {
		return nil, errors.New("wrong length after splitting at ;")
	}

	var instrumentationKey string
	var ingestionEndpoint string
	for _, part := range parts {
		subParts := strings.Split(part, "=")
		if len(subParts) != 2 {
			return nil, errors.New("wrong length after splitting at =")
		}
		switch subParts[0] {
		case "InstrumentationKey":
			instrumentationKey = subParts[1]
		case "IngestionEndpoint":
			ingestionEndpoint = subParts[1]
		default:
			return nil, fmt.Errorf("unknown attribute: %s", subParts[0])
		}
	}

	if instrumentationKey == "" {
		return nil, errors.New("missing InstrumentationKey")
	}
	if ingestionEndpoint == "" {
		return nil, errors.New("missing IngestionEndpoint")
	}

	cfg := appinsights.NewTelemetryConfiguration(instrumentationKey)
	// cfg.EndpointUrl = ingestionEndpoint
	return cfg, nil
}

func ClientFromConnectionString(connectionString string) (*appinsights.TelemetryClient, error) {
	cfg, err := ParseConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	client := appinsights.NewTelemetryClientFromConfig(cfg)
	return &client, nil
}
