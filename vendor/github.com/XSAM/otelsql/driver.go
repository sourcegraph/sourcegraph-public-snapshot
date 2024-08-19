// Copyright Sam Xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelsql

import "database/sql/driver"

var (
	_ driver.Driver        = (*otDriver)(nil)
	_ driver.DriverContext = (*otDriver)(nil)
)

type otDriver struct {
	driver driver.Driver
	cfg    config
}

func newDriver(dri driver.Driver, cfg config) driver.Driver {
	if _, ok := dri.(driver.DriverContext); ok {
		return newOtDriver(dri, cfg)
	}
	// Only implements driver.Driver
	return struct{ driver.Driver }{newOtDriver(dri, cfg)}
}

func newOtDriver(dri driver.Driver, cfg config) *otDriver {
	return &otDriver{driver: dri, cfg: cfg}
}

func (d *otDriver) Open(name string) (driver.Conn, error) {
	rawConn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return newConn(rawConn, d.cfg), nil
}

func (d *otDriver) OpenConnector(name string) (driver.Connector, error) {
	rawConnector, err := d.driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return newConnector(rawConnector, d), err
}
