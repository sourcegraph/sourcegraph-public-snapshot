import React, { Fragment, useMemo } from 'react'
import { scanSearchQuery } from '../../../shared/src/search/parser/scanner'

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<{ query: string }> = ({ query }) => {
    const tokens = useMemo(() => {
        const scannedQuery = scanSearchQuery(query)
        return scannedQuery.type === 'success'
            ? scannedQuery.term.map(token => {
                  if (token.type === 'filter') {
                      return (
                          <Fragment key={token.range.start}>
                              <span className="search-filter-keyword">
                                  {query.slice(token.field.range.start, token.field.range.end)}:
                              </span>
                              {token.value ? <>{query.slice(token.value.range.start, token.value.range.end)}</> : null}
                          </Fragment>
                      )
                  }
                  if (token.type === 'keyword') {
                      return (
                          <span className="search-keyword" key={token.range.start}>
                              {query.slice(token.range.start, token.range.end)}
                          </span>
                      )
                  }
                  return <Fragment key={token.range.start}>{query.slice(token.range.start, token.range.end)}</Fragment>
              })
            : [<Fragment key="0">{query}</Fragment>]
    }, [query])

    return <span className="text-monospace search-query-link">{tokens}</span>
}
