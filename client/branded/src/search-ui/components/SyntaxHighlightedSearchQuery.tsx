import React, { Fragment, useMemo } from 'react'

import classNames from 'classnames'

import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { decorateQuery } from '../util/query'

interface SyntaxHighlightedSearchQueryProps extends React.HTMLAttributes<HTMLSpanElement> {
    query: string
    searchPatternType?: SearchPatternType
}

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<
    React.PropsWithChildren<SyntaxHighlightedSearchQueryProps>
> = ({ query, searchPatternType, ...otherProps }) => {
    const tokens = useMemo(() => {
        const decorations = decorateQuery(query, searchPatternType)

        return decorations
            ? decorations.map(({ value, key, className }) => (
                  <span className={className} key={key}>
                      {value}
                  </span>
              ))
            : [<Fragment key="0">{query}</Fragment>]
    }, [query, searchPatternType])

    return (
        <span {...otherProps} className={classNames('text-monospace search-query-link', otherProps.className)}>
            {tokens}
        </span>
    )
}
