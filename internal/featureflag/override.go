package featureflag

import (
	"context"
	"net/http"
	"strings"
)

const (
	overrideHeader        = "X-Sourcegraph-Override-Feature"
	overrideQuery         = "feat"
	overrideQueryContains = overrideQuery + "="
)

// requestWantsTrace returns true if a request is opting into tracing either
// via our HTTP Header or our URL Query.
func requestOverrides(r *http.Request) (flags map[string]bool, ok bool) {
	// Prefer header over query param.
	values := r.Header.Values(overrideHeader)

	// PERF: Avoid parsing RawQuery if "feat=" is not present.
	if len(values) == 0 && strings.Contains(r.URL.RawQuery, overrideQueryContains) {
		values = r.URL.Query()[overrideQuery]
	}

	if len(values) == 0 {
		return nil, false
	}

	// We use this to make it convenient to specify multiple feature flag
	// overrides in many different ways. eg a user doesn't have to do multiple
	// &feat= query params, instead they could separate by space and comma.
	values = flatMapValues(values)

	flags = make(map[string]bool, len(values))
	for _, k := range values {
		// The web application uses the '~' prefix to indicate that the
		// feature flag should be reset. We need to ignore such values.
		if !strings.HasPrefix(k, "~") {
			// flags starting with "-" override to false
			v := !strings.HasPrefix(k, "-")
			k = strings.TrimPrefix(k, "-")
			flags[k] = v
		}
	}

	return flags, true
}

// overrideStore will override the returned feature flags in memory.
//
// Note: this is different to overrides in the feature flag DB, which persists
// overrides. This is intended to override feature flags for a request.
type overrideStore struct {
	store Store
	flags map[string]bool
}

func (s *overrideStore) GetUserFlags(ctx context.Context, userID int32) (map[string]bool, error) {
	return s.override(s.store.GetUserFlags(ctx, userID))
}

func (s *overrideStore) GetAnonymousUserFlags(ctx context.Context, anonUID string) (map[string]bool, error) {
	return s.override(s.store.GetAnonymousUserFlags(ctx, anonUID))
}
func (s *overrideStore) GetGlobalFeatureFlags(ctx context.Context) (map[string]bool, error) {
	return s.override(s.store.GetGlobalFeatureFlags(ctx))
}

func (s *overrideStore) override(flags map[string]bool, err error) (map[string]bool, error) {
	if err != nil {
		return nil, err
	}

	// Avoid mutating flags just in case s.store returns a cached copy.
	override := make(map[string]bool, len(flags))
	for k, v := range flags {
		override[k] = v
	}

	// Now apply overrides potentially adding or updating feature flags.
	for k, v := range s.flags {
		override[k] = v
	}

	return override, nil
}

// flatMapValues splits each string in vs by space and commas, then returns
// the flattened result.
func flatMapValues(vs []string) []string {
	var flattened []string
	for _, v := range vs {
		flattened = append(flattened, strings.FieldsFunc(v, func(r rune) bool {
			return r == ' ' || r == ','
		})...)
	}
	return flattened
}
