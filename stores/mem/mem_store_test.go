package mem_test

import (
	"testing"

	"github.com/mickamy/go-event-sourcing"
	"github.com/mickamy/go-event-sourcing/internal/storetest"
	"github.com/mickamy/go-event-sourcing/stores/mem"
)

func TestStore_Compliance(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T) ges.EventStore {
		t.Helper()
		return mem.New()
	})
}
