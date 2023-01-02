package ping_test

import (
	"testing"

	"github.com/dakotalillie/webping-lambda/internal/ping"
	"github.com/stretchr/testify/assert"
)

func TestHasTransitionedIntoErrorState(t *testing.T) {
	t.Run("not enough records", func(t *testing.T) {
		records := make([]ping.QueryRecord, 0)
		actual := ping.HasTransitionedIntoErrorState(records)
		assert.Equal(t, false, actual)
	})

	t.Run("enough records, no successes", func(t *testing.T) {
		records := []ping.QueryRecord{
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultFail},
		}
		actual := ping.HasTransitionedIntoErrorState(records)
		assert.Equal(t, true, actual)
	})

	t.Run("not enough failures", func(t *testing.T) {
		records := []ping.QueryRecord{
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultPass},
			{Result: ping.QueryResultPass},
		}
		actual := ping.HasTransitionedIntoErrorState(records)
		assert.Equal(t, false, actual)
	})

	t.Run("too many failures", func(t *testing.T) {
		records := []ping.QueryRecord{
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultFail},
		}
		actual := ping.HasTransitionedIntoErrorState(records)
		assert.Equal(t, false, actual)
	})

	t.Run("has transitioned", func(t *testing.T) {
		records := []ping.QueryRecord{
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultFail},
			{Result: ping.QueryResultPass},
		}
		actual := ping.HasTransitionedIntoErrorState(records)
		assert.Equal(t, true, actual)
	})
}
