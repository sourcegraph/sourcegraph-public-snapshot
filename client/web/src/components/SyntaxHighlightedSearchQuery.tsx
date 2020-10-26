import React from 'react'
import { parseSearchQuery } from '../../../shared/src/search/parser/parser'

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<{ query: string }> = ({ query }) => {
    const tokens = useMemo(() => {
        const parsedQuery = parseSearchQuery(query)
        return parsedQuery.type === 'success'
            ? parsedQuery.token.members.map(({ token, range }) => {
                  if (token.type === 'filter') {
                      return (
                          <>
                              <span className="search-keyword">
                                  {query.slice(token.filterType.range.start, token.filterType.range.end)}:
                              </span>
                              {token.filterValue ? (
                                  <>{query.slice(token.filterValue.range.start, token.filterValue.range.end)}</>
                              ) : null}
                          </>
                      )
                  }
                  if (token.type === 'operator') {
                      return <span className="search-operator">{query.slice(range.start, range.end)}</span>
                  }
                  return <>{query.slice(range.start, range.end)}</>
              })
            : [<>{query}</>]
    }, [query])

    return <span className="text-monospace search-query-link">{tokens}</span>
}
