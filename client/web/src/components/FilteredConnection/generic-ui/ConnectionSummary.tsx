import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Connection } from '../ConnectionType'

interface ConnectionSummaryProps<C extends Connection<N>, N> {
    // Generic connection
    connection: C
    totalCount?: number | null
    noun: string
    pluralNoun: string
    connectionQuery?: string
    /** @deprecated Required for interopability with FilteredConnection. */
    emptyElement?: JSX.Element
}

export const ConnectionSummary = <C extends Connection<N>, N>({
    connection,
    noun,
    pluralNoun,
    connectionQuery,
    totalCount,
    emptyElement,
}: ConnectionSummaryProps<C, N>): JSX.Element | null => {
    if (totalCount && totalCount > 0) {
        return (
            <p className="filtered-connection__summary" data-testid="summary">
                <small>
                    <span>
                        {totalCount} {pluralize(noun, totalCount, pluralNoun)}{' '}
                        {connectionQuery ? (
                            <span>
                                {' '}
                                matching <strong>{connectionQuery}</strong>
                            </span>
                        ) : (
                            'total'
                        )}
                    </span>{' '}
                    {connection.nodes.length < totalCount && `(showing first ${connection.nodes.length})`}
                </small>
            </p>
        )
    }

    if (connection.pageInfo?.hasNextPage) {
        // No total count to show, but it will show a 'Show more' button.
        return null
    }

    return (
        emptyElement || (
            <p className="filtered-connection__summary" data-testid="summary">
                <small>
                    No {pluralNoun}{' '}
                    {connectionQuery && (
                        <span>
                            matching <strong>{connectionQuery}</strong>
                        </span>
                    )}
                </small>
            </p>
        )
    )
}
