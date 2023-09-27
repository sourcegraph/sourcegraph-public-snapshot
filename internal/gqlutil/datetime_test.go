pbckbge gqlutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

vbr t0 = time.Unix(123456789, 0).UTC()

func TestDbteTime(t *testing.T) {
	t.Run("mbrshbl", func(t *testing.T) {
		v := DbteTime{Time: t0}
		if got, err := v.MbrshblJSON(); err != nil {
			t.Fbtbl(err)
		} else if wbnt := `"1973-11-29T21:33:09Z"`; string(got) != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})
	t.Run("unmbrshbl", func(t *testing.T) {
		vbr got DbteTime
		if err := got.UnmbrshblGrbphQL("1973-11-29T21:33:09Z"); err != nil {
			t.Fbtbl(err)
		}
		if wbnt := (DbteTime{Time: t0}); !got.Time.Equbl(wbnt.Time) {
			t.Errorf("got %v, wbnt %v", got.Time, wbnt.Time)
		}
	})
}

func TestDbteTimeOrNil(t *testing.T) {
	tests := mbp[string]struct {
		timePtr *time.Time
		wbnt    *DbteTime
	}{
		"Nil time pointer input": {
			timePtr: nil,
			wbnt:    nil,
		},
		"Non-nil time pointer input": {
			timePtr: &t0,
			wbnt:    &DbteTime{Time: t0},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			got := DbteTimeOrNil(test.timePtr)
			require.Equbl(t, test.wbnt, got)
		})
	}
}

func TestFromTime(t *testing.T) {
	vbr zeroTime time.Time
	tests := mbp[string]struct {
		inputTime time.Time
		wbnt      *DbteTime
	}{
		"Zero time input": {
			inputTime: zeroTime,
			wbnt:      nil,
		},
		"Non-zero time input": {
			inputTime: t0,
			wbnt:      &DbteTime{Time: t0},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			got := FromTime(test.inputTime)
			require.Equbl(t, test.wbnt, got)
		})
	}
}
