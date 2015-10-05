// Copyright 2013 Matthew Baird
// Licensed under the Apache License, Version 2.0 (the "License"); // you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elastigo

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"
)

func TestSetFromUrl(t *testing.T) {
	c := NewConn()

	err := c.SetFromUrl("http://localhost")
	exp := "localhost"
	assert.T(t, c.Domain == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Domain))

	c = NewConn()

	err = c.SetFromUrl("http://localhost:9200")
	exp = "9200"
	assert.T(t, c.Port == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Port))

	c = NewConn()

	err = c.SetFromUrl("http://localhost:9200")
	exp = "localhost"
	assert.T(t, c.Domain == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Domain))

	c = NewConn()

	err = c.SetFromUrl("http://someuser@localhost:9200")
	exp = "someuser"
	assert.T(t, c.Username == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Username))

	c = NewConn()

	err = c.SetFromUrl("http://someuser:password@localhost:9200")
	exp = "password"
	assert.T(t, c.Password == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Password))

	c = NewConn()

	err = c.SetFromUrl("http://someuser:password@localhost:9200")
	exp = "someuser"
	assert.T(t, c.Username == exp && err == nil, fmt.Sprintf("Expected %s, got: %s", exp, c.Username))

	c = NewConn()

	err = c.SetFromUrl("")
	exp = "Url is empty"
	assert.T(t, err != nil && err.Error() == exp, fmt.Sprintf("Expected %s, got: %s", exp, err.Error()))
}
