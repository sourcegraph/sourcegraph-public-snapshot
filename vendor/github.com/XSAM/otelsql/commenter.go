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

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type commentCarrier []string

var _ propagation.TextMapCarrier = (*commentCarrier)(nil)

func (c *commentCarrier) Keys() []string { return nil }

func (c *commentCarrier) Get(string) string { return "" }

func (c *commentCarrier) Set(key, value string) {
	*c = append(*c, fmt.Sprintf("%s='%s'", url.QueryEscape(key), url.QueryEscape(value)))
}

func (c *commentCarrier) Marshal() string {
	return strings.Join(*c, ",")
}

type commenter struct {
	enabled    bool
	propagator propagation.TextMapPropagator
}

func newCommenter(enabled bool) *commenter {
	return &commenter{
		enabled:    enabled,
		propagator: otel.GetTextMapPropagator(),
	}
}

func (c *commenter) withComment(ctx context.Context, query string) string {
	if !c.enabled {
		return query
	}

	var cc commentCarrier
	c.propagator.Inject(ctx, &cc)

	if len(cc) == 0 {
		return query
	}
	return fmt.Sprintf("%s /*%s*/", query, cc.Marshal())
}
