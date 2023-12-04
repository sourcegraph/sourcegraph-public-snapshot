package main

import (
	"testing"
	"time"
)

func TestCheckForCommit(t *testing.T) {
	tt := []struct {
		Name      string
		Version   string
		CommitLog []Commit
		Result    bool
	}{
		{
			Name:      "CommitNotFound",
			Version:   "1234567",
			CommitLog: []Commit{{Sha: "7025d04"}, {Sha: "1bbdfb1"}, {Sha: "bfe1b89"}},
			Result:    false,
		},
		{
			Name:      "CommitFound",
			Version:   "7025d04",
			CommitLog: []Commit{{Sha: "7025d04"}, {Sha: "1bbdfb1"}, {Sha: "bfe1b89"}},
			Result:    true,
		},
	}

	for _, test := range tt {
		t.Run(test.Name, func(t *testing.T) {
			result := checkForCommit(test.Version, test.CommitLog)
			if result != test.Result {
				t.Errorf("Invalid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}

func TestGetCommitFromLiveVersion(t *testing.T) {
	tt := []struct {
		Name        string
		LiveVersion string
		Result      string
	}{
		{
			Name:        "CommitFromLiveVersionSuccess",
			LiveVersion: "203800_2023-03-06_4.5-adc006d905fe",
			Result:      "adc006d905fe",
		},
		{
			Name:        "CommitFromLiveVersionFailure",
			LiveVersion: "203800_2023-03-06_4.5-confusion-adc006d905fe",
			Result:      "4.5-confusion-adc006d905fe",
		},
	}

	for _, test := range tt {
		t.Run(test.Name, func(t *testing.T) {
			result, _ := getCommitFromLiveVersion(test.LiveVersion)
			if result != test.Result {
				t.Errorf("Invalid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}

func duration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

func TestCommitTooOld(t *testing.T) {

	tt := []struct {
		Name      string
		Current   Commit
		Tip       Commit
		Threshold time.Duration
		Result    bool
	}{
		{
			Name:      "CommitWithinTime",
			Current:   Commit{Date: time.Now()},
			Tip:       Commit{Date: time.Now().Add(time.Hour * 1)},
			Threshold: duration("2h"),
			Result:    false,
		},
		{
			Name:      "CommitTooOld",
			Current:   Commit{Date: time.Now()},
			Tip:       Commit{Date: time.Now().Add(time.Hour * 3)},
			Threshold: duration("2h"),

			Result: true,
		},
		{
			Name:      "CommitWithinTimecCompoundDuration",
			Current:   Commit{Date: time.Now()},
			Tip:       Commit{Date: time.Now().Add(time.Minute * 121)},
			Threshold: duration("2h1m"),

			Result: true,
		},
		{
			Name:      "CommitOneSecondTooOld",
			Current:   Commit{Date: time.Now()},
			Tip:       Commit{Date: time.Now().Add(time.Hour * 2).Add(time.Second * 1)},
			Threshold: duration("2h1s"),

			Result: true,
		},
	}

	for _, test := range tt {
		t.Run(test.Name, func(t *testing.T) {
			result, _ := commitTooOld(test.Current, test.Tip, test.Threshold)
			if result != test.Result {
				t.Errorf("Invalid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}
