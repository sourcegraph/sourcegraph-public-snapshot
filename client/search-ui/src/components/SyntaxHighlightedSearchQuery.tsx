import React, { Fragment, useMemo } from 'react'

import classNames from 'classnames'

import { decorate, DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

interface SyntaxHighlightedSearchQueryProps extends React.HTMLAttributes<HTMLSpanElement> {
    query: string
}

interface decoration {
    value: string
    key: number
    className: string
}

function toDecoration(query: string, token: DecoratedToken): decoration {
    switch (token.type) {
        case 'field':
            return {
                value: token.value,
                key: token.range.start,
                className: 'search-filter-keyword',
            }
        case 'keyword':
            return {
                value: token.value,
                key: token.range.start,
                className: 'search-keyword',
            }
        case 'openingParen':
            return {
                value: '(',
                key: token.range.start,
                className: 'search-keyword',
            }
        case 'closingParen':
            return {
                value: ')',
                key: token.range.start,
                className: 'search-keyword',
            }

        case 'metaFilterSeparator':
            return {
                value: ':',
                key: token.range.start,
                className: 'search-filter-separator',
            }
        case 'metaRepoRevisionSeparator':
        case 'metaContextPrefix':
            return {
                value: '@',
                key: token.range.start,
                className: 'search-keyword',
            }
        case 'metaPath':
            return {
                value: token.value,
                key: token.range.start,
                className: 'search-path-separator',
            }

        case 'metaRevision': {
            let kind = ''
            switch (token.kind) {
                case 'Separator':
                    kind = 'separator'
                    break
                case 'IncludeGlobMarker':
                    kind = 'include-glob-marker'
                    break
                case 'ExcludeGlobMarker':
                    kind = 'exclude-glob-marker'
                    break
                case 'CommitHash':
                    kind = 'commit-hash'
                    break
                case 'Label':
                    kind = 'label'
                    break
                case 'ReferencePath':
                    kind = 'reference-path'
                    break
                case 'Wildcard':
                    kind = 'wildcard'
                    break
            }
            return {
                value: token.value,
                key: token.range.start,
                className: `search-revision-${kind}`,
            }
        }

        case 'metaRegexp': {
            let kind = ''
            switch (token.kind) {
                case 'Assertion':
                    kind = 'assertion'
                    break
                case 'Alternative':
                    kind = 'alternative'
                    break
                case 'Delimited':
                    kind = 'delimited'
                    break
                case 'EscapedCharacter':
                    kind = 'escaped-character'
                    break
                case 'CharacterSet':
                    kind = 'character-set'
                    break
                case 'CharacterClass':
                    kind = 'character-class'
                    break
                case 'CharacterClassRange':
                    kind = 'character-class-range'
                    break
                case 'CharacterClassRangeHyphen':
                    kind = 'character-class-range-hyphen'
                    break
                case 'CharacterClassMember':
                    kind = 'character-class-member'
                    break
                case 'LazyQuantifier':
                    kind = 'lazy-quantifier'
                    break
                case 'RangeQuantifier':
                    kind = 'range-quantifier'
                    break
            }
            return {
                value: token.value,
                key: token.range.start,
                className: `search-regexp-meta-${kind}`,
            }
        }

        case 'metaPredicate': {
            let kind = ''
            let value = ''
            switch (token.kind) {
                case 'NameAccess':
                    kind = 'name-access'
                    value = query.slice(token.range.start, token.range.end)
                    break
                case 'Dot':
                    kind = 'dot'
                    value = '.'
                    break
                case 'Parenthesis':
                    kind = 'parenthesis'
                    value = query.slice(token.range.start, token.range.end)
                    break
            }
            return {
                value,
                key: token.range.start,
                className: `search-predicate-${kind}`,
            }
        }

        case 'metaStructural': {
            let kind = ''
            switch (token.kind) {
                case 'Hole':
                    kind = 'hole'
                    break
                case 'RegexpHole':
                    kind = 'regexp-hole'
                    break
                case 'Variable':
                    kind = 'variable'
                    break
                case 'RegexpSeparator':
                    kind = 'regexp-separator'
                    break
            }
            return {
                value: token.value,
                key: token.range.start,
                className: `search-structural-${kind}`,
            }
        }
    }
    return {
        value: query.slice(token.range.start, token.range.end),
        key: token.range.start,
        className: 'search-query-text',
    }
}

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<
    React.PropsWithChildren<SyntaxHighlightedSearchQueryProps>
> = ({ query, ...otherProps }) => {
    const tokens = useMemo(() => {
        const tokens = scanSearchQuery(query)
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
    }, [query])

    return (
        <span {...otherProps} className={classNames('text-monospace search-query-link', otherProps.className)}>
            {tokens}
        </span>
    )
}
