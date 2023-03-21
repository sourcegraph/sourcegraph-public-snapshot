import React, { Fragment, useMemo } from 'react'

import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { decorate, toDecoration } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

import { HighlightedLabel } from './Suggestions'

interface SyntaxHighlightedSearchQueryProps extends React.HTMLAttributes<HTMLSpanElement> {
    query: string
    matches?: Set<number>
    searchPatternType?: SearchPatternType
}

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<
    React.PropsWithChildren<SyntaxHighlightedSearchQueryProps>
> = ({ query, searchPatternType, matches, ...otherProps }) => {
    const tokens = useMemo(() => {
        const tokens = searchPatternType ? scanSearchQuery(query, false, searchPatternType) : scanSearchQuery(query)
        return tokens.type === 'success'
            ? tokens.term.flatMap(token =>
                  decorate(token).map(token => {
                      const { value, key, className } = toDecoration(query, token)
                      return (
                          <span className={className} key={key}>
                              {matches ? (
                                  <HighlightedLabel label={value} matches={matches} offset={token.range.start} />
                              ) : (
                                  value
                              )}
                          </span>
                      )
                  })
              )
            : [<Fragment key="0">{query}</Fragment>]
    }, [query, matches, searchPatternType])

    return (
        <span {...otherProps} className={classNames('text-monospace search-query-link', otherProps.className)}>
            {tokens}
        </span>
    )
}
