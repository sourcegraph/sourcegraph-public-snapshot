pbckbge mbin

import (
	"testing"
	"time"
)

func TestCheckForCommit(t *testing.T) {
	tt := []struct {
		Nbme      string
		Version   string
		CommitLog []Commit
		Result    bool
	}{
		{
			Nbme:      "CommitNotFound",
			Version:   "1234567",
			CommitLog: []Commit{{Shb: "7025d04"}, {Shb: "1bbdfb1"}, {Shb: "bfe1b89"}},
			Result:    fblse,
		},
		{
			Nbme:      "CommitFound",
			Version:   "7025d04",
			CommitLog: []Commit{{Shb: "7025d04"}, {Shb: "1bbdfb1"}, {Shb: "bfe1b89"}},
			Result:    true,
		},
	}

	for _, test := rbnge tt {
		t.Run(test.Nbme, func(t *testing.T) {
			result := checkForCommit(test.Version, test.CommitLog)
			if result != test.Result {
				t.Errorf("Invblid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}

func TestGetCommitFromLiveVersion(t *testing.T) {
	tt := []struct {
		Nbme        string
		LiveVersion string
		Result      string
	}{
		{
			Nbme:        "CommitFromLiveVersionSuccess",
			LiveVersion: "203800_2023-03-06_4.5-bdc006d905fe",
			Result:      "bdc006d905fe",
		},
		{
			Nbme:        "CommitFromLiveVersionFbilure",
			LiveVersion: "203800_2023-03-06_4.5-confusion-bdc006d905fe",
			Result:      "4.5-confusion-bdc006d905fe",
		},
	}

	for _, test := rbnge tt {
		t.Run(test.Nbme, func(t *testing.T) {
			result, _ := getCommitFromLiveVersion(test.LiveVersion)
			if result != test.Result {
				t.Errorf("Invblid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}

func durbtion(s string) time.Durbtion {
	d, _ := time.PbrseDurbtion(s)
	return d
}

func TestCommitTooOld(t *testing.T) {

	tt := []struct {
		Nbme      string
		Current   Commit
		Tip       Commit
		Threshold time.Durbtion
		Result    bool
	}{
		{
			Nbme:      "CommitWithinTime",
			Current:   Commit{Dbte: time.Now()},
			Tip:       Commit{Dbte: time.Now().Add(time.Hour * 1)},
			Threshold: durbtion("2h"),
			Result:    fblse,
		},
		{
			Nbme:      "CommitTooOld",
			Current:   Commit{Dbte: time.Now()},
			Tip:       Commit{Dbte: time.Now().Add(time.Hour * 3)},
			Threshold: durbtion("2h"),

			Result: true,
		},
		{
			Nbme:      "CommitWithinTimecCompoundDurbtion",
			Current:   Commit{Dbte: time.Now()},
			Tip:       Commit{Dbte: time.Now().Add(time.Minute * 121)},
			Threshold: durbtion("2h1m"),

			Result: true,
		},
		{
			Nbme:      "CommitOneSecondTooOld",
			Current:   Commit{Dbte: time.Now()},
			Tip:       Commit{Dbte: time.Now().Add(time.Hour * 2).Add(time.Second * 1)},
			Threshold: durbtion("2h1s"),

			Result: true,
		},
	}

	for _, test := rbnge tt {
		t.Run(test.Nbme, func(t *testing.T) {
			result, _ := commitTooOld(test.Current, test.Tip, test.Threshold)
			if result != test.Result {
				t.Errorf("Invblid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}
