import React, { Fragment, useMemo } from 'react'

import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/search'
import { decorate, DecoratedToken, toCSSClassName } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

interface SyntaxHighlightedSearchQueryProps extends React.HTMLAttributes<HTMLSpanElement> {
    query: string
    searchPatternType?: SearchPatternType
}

interface decoration {
    value: string
    key: number
    className: string
}

function toDecoration(query: string, token: DecoratedToken): decoration {
    const className = toCSSClassName(token)

    switch (token.type) {
        case 'keyword':
        case 'field':
        case 'metaPath':
        case 'metaRevision':
        case 'metaRegexp':
        case 'metaStructural':
            return {
                value: token.value,
                key: token.range.start,
                className,
            }
        case 'openingParen':
            return {
                value: '(',
                key: token.range.start,
                className,
            }
        case 'closingParen':
            return {
                value: ')',
                key: token.range.start,
                className,
            }

        case 'metaFilterSeparator':
            return {
                value: ':',
                key: token.range.start,
                className,
            }
        case 'metaRepoRevisionSeparator':
        case 'metaContextPrefix':
            return {
                value: '@',
                key: token.range.start,
                className,
            }

        case 'metaPredicate': {
            let value = ''
            switch (token.kind) {
                case 'NameAccess':
                    value = query.slice(token.range.start, token.range.end)
                    break
                case 'Dot':
                    value = '.'
                    break
                case 'Parenthesis':
                    value = query.slice(token.range.start, token.range.end)
                    break
            }
            return {
                value,
                key: token.range.start,
                className,
            }
        }
    }
    return {
        value: query.slice(token.range.start, token.range.end),
        key: token.range.start,
        className,
    }
}

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<
    React.PropsWithChildren<SyntaxHighlightedSearchQueryProps>
> = ({ query, searchPatternType, ...otherProps }) => {
    const tokens = useMemo(() => {
        const tokens = searchPatternType ? scanSearchQuery(query) : scanSearchQuery(query, false, searchPatternType)
        return tokens.type === 'success'
            ? tokens.term.flatMap(token =>
                  decorate(token).map(token => {
                      const { value, key, className } = toDecoration(query, token)
                      return (
                          <span className={className} key={key}>
                              {value}
                          </span>
                      )
                  })
              )
            : [<Fragment key="0">{query}</Fragment>]
    }, [query, searchPatternType])

    return (
        <span {...otherProps} className={classNames('text-monospace search-query-link', otherProps.className)}>
            {tokens}
        </span>
    )
}
