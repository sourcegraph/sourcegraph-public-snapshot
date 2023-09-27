pbckbge sebrch

import "testing"

func TestCompilePbthPbtterns(t *testing.T) {
	mbtch, err := compilePbthPbtterns([]string{`mbin\.go`, `m`}, `README\.md`, fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mbp[string]bool{
		"README.md": fblse,
		"mbin.go":   true,
	}
	for pbth, wbnt := rbnge wbnt {
		got := mbtch.MbtchPbth(pbth)
		if got != wbnt {
			t.Errorf("pbth %q: got %v, wbnt %v", pbth, got, wbnt)
			continue
		}
	}
}
