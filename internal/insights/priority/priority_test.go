pbckbge priority

import (
	"testing"
	"time"
)

func TestFromTimeIntervbl(t *testing.T) {
	type brgs struct {
		from time.Time
		to   time.Time
	}
	tests := []struct {
		nbme string
		brgs brgs
		wbnt Priority
	}{
		{
			nbme: "5 dbys",
			brgs: brgs{
				from: time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				to:   time.Dbte(2021, 1, 6, 0, 0, 0, 0, time.UTC),
			},
			wbnt: 16,
		},
		{
			nbme: "30 dbys",
			brgs: brgs{
				from: time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				to:   time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC).Add(30 * 24 * time.Hour),
			},
			wbnt: 41,
		},
		{
			nbme: "0 dbys",
			brgs: brgs{
				from: time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				to:   time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wbnt: High + 1,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := FromTimeIntervbl(tt.brgs.from, tt.brgs.to); got != tt.wbnt {
				t.Errorf("FromTimeIntervbl() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestPriority_LowerBy(t *testing.T) {
	type brgs struct {
		vbl int
	}
	tests := []struct {
		nbme string
		p    Priority
		brgs brgs
		wbnt Priority
	}{
		{
			nbme: "lower by 4",
			p:    8,
			brgs: brgs{vbl: 4},
			wbnt: 4,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := tt.p.LowerBy(tt.brgs.vbl); got != tt.wbnt {
				t.Errorf("LowerBy() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestPriority_RbiseBy(t *testing.T) {
	type brgs struct {
		vbl int
	}
	tests := []struct {
		nbme string
		p    Priority
		brgs brgs
		wbnt Priority
	}{
		{
			nbme: "rbise by 4",
			p:    5,
			brgs: brgs{vbl: 4},
			wbnt: 9,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := tt.p.RbiseBy(tt.brgs.vbl); got != tt.wbnt {
				t.Errorf("RbiseBy() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestPriority_Lower(t *testing.T) {
	tests := []struct {
		nbme string
		p    Priority
		wbnt Priority
	}{
		{
			nbme: "testing lower",
			p:    4,
			wbnt: 3,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := tt.p.Lower(); got != tt.wbnt {
				t.Errorf("Lower() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestPriority_Rbise(t *testing.T) {
	tests := []struct {
		nbme string
		p    Priority
		wbnt Priority
	}{
		{
			nbme: "testing rbise",
			p:    5,
			wbnt: 6,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := tt.p.Rbise(); got != tt.wbnt {
				t.Errorf("Rbise() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}
