import React from 'react'
import { parseSearchQuery } from '../../../shared/src/search/parser/parser'

// A read-only syntax highlighted search query
export const SyntaxHighlightedSearchQuery: React.FunctionComponent<{ query: string }> = ({ query }) => {
    const parsedQuery = parseSearchQuery(query)

    const tokens: JSX.Element[] = []

    if (parsedQuery.type === 'success') {
        for (const member of parsedQuery.token.members) {
            if (member.token.type === 'filter') {
                tokens.push(
                    <>
                        <span className="search-keyword">
                            {query.slice(member.token.filterType.range.start, member.token.filterType.range.end)}:
                        </span>
                        {member.token.filterValue ? (
                            <>{query.slice(member.token.filterValue.range.start, member.token.filterValue.range.end)}</>
                        ) : null}
                    </>
                )
            } else if (member.token.type === 'operator') {
                tokens.push(
                    <span className="search-operator">{query.slice(member.range.start, member.range.end)}</span>
                )
            } else {
                tokens.push(<>{query.slice(member.range.start, member.range.end)}</>)
            }
        }
    } else {
        tokens.push(<>{query}</>)
    }

    return <span className="text-monospace search-query-link">{tokens}</span>
}
