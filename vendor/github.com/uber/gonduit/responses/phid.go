package responses

import "github.com/uber/gonduit/entities"

// PHIDQueryResponse is the result of phid.query operations.
type PHIDQueryResponse map[string]*entities.PHIDResult

// PHIDLookupResponse is the result of phid.lookup operations.
type PHIDLookupResponse map[string]*entities.PHIDResult
