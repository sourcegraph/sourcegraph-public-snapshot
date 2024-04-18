package honey

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"reflect"
	"sync"
	"testing"
)

func TestNonSendingEvent(t *testing.T) {
	t.Run("AddField and Fields", func(t *testing.T) {
		event := NonSendingEvent()

		// Add some fields
		event.AddField("key1", "value1")
		event.AddField("key2", 42)
		event.AddField("key3", true)

		// Check if the fields were added correctly
		expectedFields := map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		assert.EqualValues(t, event.Fields(), expectedFields)
	})

	t.Run("AddField is thread-safe", func(t *testing.T) {
		event := NonSendingEvent()

		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				event.AddField(fmt.Sprintf("key%d", i), i)
			}(i)
		}
		wg.Wait()

		// Check if all fields were added correctly
		if len(event.Fields()) != 1000 {
			t.Errorf("Expected 1000 fields, got %d", len(event.Fields()))
		}
	})

	t.Run("AddAttributes", func(t *testing.T) {
		event := NonSendingEvent()

		// Add some attributes
		attrs := []attribute.KeyValue{
			attribute.String("key1", "value1"),
			attribute.Int64("key2", 42),
			attribute.Bool("key3", true),
		}
		event.AddAttributes(attrs)

		// Check if the attributes were added correctly
		expectedFields := map[string]any{
			"key1": "value1",
			"key2": int64(42),
			"key3": true,
		}

		if !reflect.DeepEqual(event.Fields(), expectedFields) {
			t.Errorf("Fields mismatch. Expected: %v, Got: %v", expectedFields, event.Fields())
		}
	})

	t.Run("AddAttributes is thread-safe", func(t *testing.T) {
		event := NonSendingEvent()

		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				attrs := []attribute.KeyValue{
					attribute.String(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i)),
				}
				event.AddAttributes(attrs)
			}(i)
		}
		wg.Wait()

		// Check if all attributes were added correctly
		if len(event.Fields()) != 1000 {
			t.Errorf("Expected 1000 fields, got %d", len(event.Fields()))
		}
	})
}
