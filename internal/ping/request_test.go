package ping_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dakotalillie/webping-lambda/internal/ping"
	"github.com/stretchr/testify/assert"
)

func TestSendRequestToEndpoint(t *testing.T) {
	t.Run("no protocol", func(t *testing.T) {
		record := ping.SendRequestToEndpoint(context.Background(), "google.com")
		assert.Equal(t, ping.QueryResultFail, record.Result)
	})

	t.Run("404 endpoint", func(t *testing.T) {
		record := ping.SendRequestToEndpoint(context.Background(), "https://google.com/404")
		assert.Equal(t, ping.QueryResultFail, record.Result)
	})

	t.Run("success", func(t *testing.T) {
		record := ping.SendRequestToEndpoint(context.Background(), "https://google.com")
		assert.Equal(t, ping.QueryResultPass, record.Result)
	})
}

func TestSendRequestsToAllEndpoints(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		records := ping.SendRequestsToAllEndpoints(context.Background(), []string{"https://google.com", "https://apple.com"})
		for _, record := range records {
			assert.Equal(
				t,
				ping.QueryResultPass,
				record.Result,
				fmt.Sprintf("received error response for %s", record.Endpoint),
			)
		}
	})
}
