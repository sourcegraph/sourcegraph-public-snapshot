import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { Text } from '@sourcegraph/wildcard'

import { type ConnectionNodesState, type ConnectionProps, getTotalCount } from '../ConnectionNodes'
import type { Connection } from '../ConnectionType'

import styles from './ConnectionSummary.module.scss'

interface ConnectionNodesSummaryProps<C extends Connection<N>, N, NP = {}, HP = {}>
    extends Pick<
        ConnectionProps<N, NP, HP> & ConnectionNodesState,
        | 'noSummaryIfAllNodesVisible'
        | 'totalCountSummaryComponent'
        | 'noun'
        | 'pluralNoun'
        | 'connectionQuery'
        | 'emptyElement'
        | 'first'
    > {
    /** The fetched connection data or an error (if an error occurred). */
    connection: C

    hasNextPage: boolean

    compact?: boolean

    centered?: boolean

    className?: string
}

/**
 * FilteredConnection summary content.
 * Used to configure a suitable summary from a specific connection response.
 */
export const ConnectionSummary = <C extends Connection<N>, N, NP = {}, HP = {}>({
    noSummaryIfAllNodesVisible,
    connection,
    hasNextPage,
    totalCountSummaryComponent: TotalCountSummaryComponent,
    noun,
    pluralNoun,
    connectionQuery,
    emptyElement,
    first,
    compact,
    centered,
    className,
}: ConnectionNodesSummaryProps<C, N, NP, HP>): JSX.Element | null => {
    const shouldShowSummary = !noSummaryIfAllNodesVisible || connection.nodes.length === 0 || hasNextPage
    const summaryClassName = classNames(
        compact && styles.compact,
        centered && styles.centered,
        styles.normal,
        className
    )

    if (!shouldShowSummary) {
        return null
    }

    // We cannot always rely on `connection.totalCount` to be returned, fallback to `connection.nodes.length` if possible.
    const totalCount = getTotalCount(connection, first)

    if (totalCount !== null && totalCount > 0 && TotalCountSummaryComponent) {
        return <TotalCountSummaryComponent totalCount={totalCount} />
    }

    if (totalCount !== null && totalCount > 0 && !TotalCountSummaryComponent) {
        return (
            <Text className={summaryClassName} data-testid="summary">
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
            </Text>
        )
    }

    if (connection.pageInfo?.hasNextPage) {
        // No total count to show, but it will show a 'Show more' button.
        return null
    }

    return (
        emptyElement || (
            <Text className={summaryClassName} data-testid="summary">
                <small>
                    No {pluralNoun}{' '}
                    {connectionQuery && (
                        <span>
                            matching <strong>{connectionQuery}</strong>
                        </span>
                    )}
                </small>
            </Text>
        )
    )
}
