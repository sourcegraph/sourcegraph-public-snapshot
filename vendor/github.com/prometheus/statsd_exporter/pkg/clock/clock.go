// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clock

import (
	"time"
)

var ClockInstance *Clock

type Clock struct {
	Instant  time.Time
	TickerCh chan time.Time
}

func Now() time.Time {
	if ClockInstance == nil {
		return time.Now()
	}
	return ClockInstance.Instant
}

func NewTicker(d time.Duration) *time.Ticker {
	if ClockInstance == nil || ClockInstance.TickerCh == nil {
		return time.NewTicker(d)
	}
	return &time.Ticker{
		C: ClockInstance.TickerCh,
	}
}
