package window

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func (t timeOfDay) Equal(other timeOfDay) bool {
	return t.cmp == other.cmp
}

func TestTimeOfDay(t *testing.T) {
	early := timeOfDayFromParts(2, 0)
	late := timeOfDayFromTime(time.Date(2021, 4, 7, 19, 37, 0, 0, time.UTC))
	alsoLate := late

	assert.True(t, early.before(late))
	assert.False(t, early.after(late))
	assert.True(t, late.after(early))
	assert.False(t, late.before(early))
	assert.True(t, alsoLate.Equal(late))
	assert.False(t, alsoLate.Equal(early))
}
