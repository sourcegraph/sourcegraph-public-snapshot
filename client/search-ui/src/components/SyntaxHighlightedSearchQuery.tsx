import React, { Fragment, useMemo } from 'react'

import classNames from 'classnames'

import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

interface SyntaxHighlightedSearchQueryProps extends React.HTMLAttributes<HTMLSpanElement> {
    query: string
}

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<
    React.PropsWithChildren<SyntaxHighlightedSearchQueryProps>
> = ({ query, ...otherProps }) => {
    const tokens = useMemo(() => {
        const scannedQuery = scanSearchQuery(query)
        return scannedQuery.type === 'success'
            ? scannedQuery.term.map(token => {
                  if (token.type === 'filter') {
                      return (
                          <Fragment key={token.range.start}>
                              <span className="search-filter-keyword">
                                  {query.slice(token.field.range.start, token.field.range.end)}
                              </span>
                              <span className="search-filter-separator">:</span>
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

    return (
        <span {...otherProps} className={classNames('text-monospace search-query-link', otherProps.className)}>
            {tokens}
        </span>
    )
}
