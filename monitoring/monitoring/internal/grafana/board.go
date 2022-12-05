package grafana

import (
	"github.com/grafana-tools/sdk"
)

// Board creates a dashboard with some default configurations.
func Board(uid, title string, labels []string) *sdk.Board {
	board := sdk.NewBoard(title)
	board.AddTags(labels...)
	board.UID = uid
	board.ID = 0
	board.Timezone = "utc"
	board.Timepicker.RefreshIntervals = []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"}
	board.Time.From = "now-6h"
	board.Time.To = "now"
	board.SharedCrosshair = true
	board.Editable = false
	return board
}
