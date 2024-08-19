// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

type ymdDesignator byte
type hmsDesignator byte

const (
	Year  ymdDesignator = 'Y'
	Month ymdDesignator = 'M'
	Week  ymdDesignator = 'W'
	Day   ymdDesignator = 'D'

	Hour   hmsDesignator = 'H'
	Minute hmsDesignator = 'M'
	Second hmsDesignator = 'S'
)

func (d ymdDesignator) IsOneOf(xx ...ymdDesignator) bool {
	for _, x := range xx {
		if x == d {
			return true
		}
	}
	return false
}

func (d ymdDesignator) IsNotOneOf(xx ...ymdDesignator) bool {
	for _, x := range xx {
		if x == d {
			return false
		}
	}
	return true
}

func (d hmsDesignator) IsOneOf(xx ...hmsDesignator) bool {
	for _, x := range xx {
		if x == d {
			return true
		}
	}
	return false
}

func (d hmsDesignator) IsNotOneOf(xx ...hmsDesignator) bool {
	for _, x := range xx {
		if x == d {
			return false
		}
	}
	return true
}
