package grafana

import (
	"github.com/grafana-tools/sdk"
)

// NewBoard creates a dashboard with some default configurations.
func NewBoard(uid, title string, labels []string) *sdk.Board {
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

// NewRowPanel creates a row, and should be used instead of the native board.AddRow because
// it seems to have better API support. The returned Panel should be added to a *sdk.Board,
// and panels in this row should be added to both the returned Panel and the parent
// *sdk.Board.
func NewRowPanel(offsetY int, title string) *sdk.Panel {
	row := &sdk.Panel{RowPanel: &sdk.RowPanel{}}
	row.OfType = sdk.RowType
	row.Type = "row"
	row.Title = title
	row.Panels = []sdk.Panel{} // cannot be null

	// set position
	zero := 0
	row.GridPos.X = &zero
	row.GridPos.Y = &offsetY

	return row
}
