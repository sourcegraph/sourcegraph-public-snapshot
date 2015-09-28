package humanize

import (
	"testing"
	"time"
)

func checkTime(t *testing.T, expected, got string) {
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestPast(t *testing.T) {

	const longAgo = 37 * 365 * int64(time.Hour) * 24

	expected := []string{
		"now",
		"1 second ago",
		"12 seconds ago",
		"30 seconds ago",
		"45 seconds ago",
		"1 minute ago",
		"15 minutes ago",
		"1 hour ago",
		"2 hours ago",
		"21 hours ago",
		"1 day ago",
		"2 days ago",
		"3 days ago",
		"1 week ago",
		"1 week ago",
		"2 weeks ago",
		"1 month ago",
		"3 months ago",
		"1 year ago",
	}

	i := 0
	now := time.Now().Unix()

	expected = append(expected, time.Unix(now-longAgo, 0).String())

	checkTime(t, expected[i], Time(time.Unix(now, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-1, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-12, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-30, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-45, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-63, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-15*Minute, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-63*Minute, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-2*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-21*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-26*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-49*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-3*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-7*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-12*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-15*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-39*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-99*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-365*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now-longAgo, 0)))
}

func TestFuture(t *testing.T) {

	const awhileFromNow = 37 * 365 * int64(time.Hour) * 24

	expected := []string{
		"now",
		"1 second from now",
		"12 seconds from now",
		"30 seconds from now",
		"45 seconds from now",
		"15 minutes from now",
		"2 hours from now",
		"21 hours from now",
		"1 day from now",
		"2 days from now",
		"3 days from now",
		"1 week from now",
		"1 week from now",
		"2 weeks from now",
		"1 month from now",
		"1 year from now",
	}

	i := 0
	now := time.Now().Unix()

	expected = append(expected, time.Unix(now+awhileFromNow, 0).String())

	checkTime(t, expected[i], Time(time.Unix(now, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+1, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+12, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+30, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+45, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+15*Minute, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+2*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+21*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+26*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+49*Hour, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+3*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+7*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+12*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+15*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+39*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+365*Day, 0)))
	i++
	checkTime(t, expected[i], Time(time.Unix(now+awhileFromNow, 0)))
}
