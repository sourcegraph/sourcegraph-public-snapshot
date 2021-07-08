import * as React from 'react'

import { ConnectionNodesState, ConnectionProps } from './ConnectionNodes'
import { Connection } from './ConnectionType'
import { ConnectionSummary } from './generic-ui'

interface ConnectionNodesSummaryProps<C extends Connection<N>, N, NP = {}, HP = {}>
    extends Pick<
        ConnectionProps<N, NP, HP> & ConnectionNodesState,
        | 'noSummaryIfAllNodesVisible'
        | 'totalCountSummaryComponent'
        | 'noun'
        | 'pluralNoun'
        | 'connectionQuery'
        | 'emptyElement'
    > {
    /** The fetched connection data or an error (if an error occurred). */
    connection: C

    hasNextPage: boolean

    totalCount: number | null
}

export const ConnectionNodesSummary = <C extends Connection<N>, N, NP = {}, HP = {}>({
    noSummaryIfAllNodesVisible,
    connection,
    hasNextPage,
    totalCount,
    totalCountSummaryComponent: TotalCountSummaryComponent,
    noun,
    pluralNoun,
    connectionQuery,
    emptyElement,
}: ConnectionNodesSummaryProps<C, N, NP, HP>): JSX.Element | null => {
    const shouldShowSummary = !noSummaryIfAllNodesVisible || connection.nodes.length === 0 || hasNextPage

    if (!shouldShowSummary) {
        return null
    }

    if (totalCount !== null && totalCount > 0 && TotalCountSummaryComponent) {
        return <TotalCountSummaryComponent totalCount={totalCount} />
    }

    return (
        <ConnectionSummary
            connection={connection}
            totalCount={totalCount}
            noun={noun}
            pluralNoun={pluralNoun}
            connectionQuery={connectionQuery}
            emptyElement={emptyElement}
        />
    )
}
