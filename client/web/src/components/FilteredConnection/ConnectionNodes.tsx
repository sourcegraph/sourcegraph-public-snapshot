import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ConnectionNodesSummary } from './ConnectionNodesSummary'
import { Connection } from './ConnectionType'
import { hasID } from './utils'

/**
 * Props for the FilteredConnection component's result nodes and associated summary/pagination controls.
 *
 * @template N The node type of the GraphQL connection, such as GQL.IRepository (if the connection is GQL.IRepositoryConnection)
 * @template NP Props passed to `nodeComponent` in addition to `{ node: N }`
 * @template HP Props passed to `headComponent` in addition to `{ nodes: N[]; totalCount?: number | null }`.
 */
export interface ConnectionProps<N, NP = {}, HP = {}> extends ConnectionNodesDisplayProps {
    /** Header row to appear above all nodes. */
    headComponent?: React.ComponentType<{ nodes: N[]; totalCount?: number | null } & HP>

    /** Props to pass to each headComponent in addition to `{ nodes: N[]; totalCount?: number | null }`. */
    headComponentProps?: HP

    /** Footer row to appear below all nodes. */
    footComponent?: React.ComponentType<{ nodes: N[] }>

    /** The component type to use to display each node. */
    nodeComponent: React.ComponentType<{ node: N } & NP>

    /** Props to pass to each nodeComponent in addition to `{ node: N }`. */
    nodeComponentProps?: NP

    /** An element rendered as a sibling of the filters. */
    additionalFilterElement?: React.ReactElement
}

/** State related to the ConnectionNodes component. */
export interface ConnectionNodesState {
    query: string
    first: number

    connectionQuery?: string

    /** The `PageInfo.endCursor` value from the previous request. */
    after?: string

    /**
     * The number of results that were visible from previous requests. The initial request of
     * a result set will load `visible` items, then will request `first` items on each subsequent
     * request. This has the effect of loading the correct number of visible results when a URL
     * is copied during pagination. This value is only useful with cursor-based paging.
     */
    visible?: number

    /**
     * Whether the connection is loading. It is not equivalent to connection === undefined because we preserve the
     * old data for ~250msec while loading to reduce jitter.
     */
    loading: boolean
}

/**
 * Fields that belong in ConnectionNodesProps and that don't depend on the type parameters. These are the fields
 * that are most likely to be needed by callers, and it's simpler for them if they are in a parameter-less type.
 */
export interface ConnectionNodesDisplayProps {
    /** list HTML element type. Default is <ul>. */
    listComponent?: 'ul' | 'table' | 'div'

    /** CSS class name for the list element (<ul>, <table>, or <div>). */
    listClassName?: string

    /** CSS class name for the "Show more" button. */
    showMoreClassName?: string

    /** The English noun (in singular form) describing what this connection contains. */
    noun: string

    /** The English noun (in plural form) describing what this connection contains. */
    pluralNoun: string

    /** Do not show a "Show more" button. */
    noShowMore?: boolean

    /** Do not show a count summary if all nodes are visible in the list's first page. */
    noSummaryIfAllNodesVisible?: boolean

    /** The component displayed when the list of nodes is empty. */
    emptyElement?: JSX.Element

    /** The component displayed when all nodes have been fetched. */
    totalCountSummaryComponent?: React.ComponentType<{ totalCount: number }>

    // TODO: Move this to FilteredConnection
    /**
     * Set to true when the GraphQL response is expected to emit an `PageInfo.endCursor` value when
     * there is a subsequent page of results. This will request the next page of results and append
     * them onto the existing list of results instead of requesting twice as many results and
     * replacing the existing results.
     */
    cursorPaging?: boolean
}

interface ConnectionNodesProps<C extends Connection<N>, N, NP = {}, HP = {}>
    extends ConnectionProps<N, NP, HP>,
        ConnectionNodesState {
    /** The fetched connection data or an error (if an error occurred). */
    connection: C

    location: H.Location

    onShowMore: () => void
}

export class ConnectionNodes<C extends Connection<N>, N, NP = {}, HP = {}> extends React.PureComponent<
    ConnectionNodesProps<C, N, NP, HP>
> {
    public render(): JSX.Element | null {
        const NodeComponent = this.props.nodeComponent
        const ListComponent = this.props.listComponent || 'ul'
        const HeadComponent = this.props.headComponent
        const FootComponent = this.props.footComponent
        const TotalCountSummaryComponent = this.props.totalCountSummaryComponent

        const hasNextPage = this.props.connection.pageInfo
            ? this.props.connection.pageInfo.hasNextPage
            : typeof this.props.connection.totalCount === 'number' &&
              this.props.connection.nodes.length < this.props.connection.totalCount

        let totalCount: number | null = null
        if (typeof this.props.connection.totalCount === 'number') {
            totalCount = this.props.connection.totalCount
        } else if (
            // TODO(sqs): this line below is wrong because this.props.first might've just been changed and
            // this.props.connection.nodes is still the data fetched from before this.props.first was changed.
            // this causes the UI to incorrectly show "N items total" even when the count is indeterminate right
            // after the user clicks "Show more" but before the new data is loaded.
            this.props.connection.nodes.length < this.props.first ||
            (this.props.connection.nodes.length === this.props.first &&
                this.props.connection.pageInfo &&
                typeof this.props.connection.pageInfo.hasNextPage === 'boolean' &&
                !this.props.connection.pageInfo.hasNextPage)
        ) {
            totalCount = this.props.connection.nodes.length
        }

        let summary: React.ReactFragment | undefined
        if (!this.props.noSummaryIfAllNodesVisible || this.props.connection.nodes.length === 0 || hasNextPage) {
            if (totalCount !== null && totalCount > 0) {
                summary = TotalCountSummaryComponent ? (
                    <TotalCountSummaryComponent totalCount={totalCount} />
                ) : (
                    <p className="filtered-connection__summary" data-testid="summary">
                        <small>
                            <span>
                                {totalCount} {pluralize(this.props.noun, totalCount, this.props.pluralNoun)}{' '}
                                {this.props.connectionQuery ? (
                                    <span>
                                        {' '}
                                        matching <strong>{this.props.connectionQuery}</strong>
                                    </span>
                                ) : (
                                    'total'
                                )}
                            </span>{' '}
                            {this.props.connection.nodes.length < totalCount &&
                                `(showing first ${this.props.connection.nodes.length})`}
                        </small>
                    </p>
                )
            } else if (this.props.connection.pageInfo?.hasNextPage) {
                // No total count to show, but it will show a 'Show more' button.
            } else if (totalCount === 0) {
                summary = this.props.emptyElement || (
                    <p className="filtered-connection__summary" data-testid="summary">
                        <small>
                            No {this.props.pluralNoun}{' '}
                            {this.props.connectionQuery && (
                                <span>
                                    matching <strong>{this.props.connectionQuery}</strong>
                                </span>
                            )}
                        </small>
                    </p>
                )
            }
        }

        const nodes = this.props.connection.nodes.map((node, index) => (
            <NodeComponent key={hasID(node) ? node.id : index} node={node} {...this.props.nodeComponentProps!} />
        ))

        return (
            <>
                {this.props.connectionQuery && <ConnectionNodesSummary summary={summary} />}
                {this.props.connection.nodes.length > 0 && (
                    <ListComponent
                        className={classNames('filtered-connection__nodes', this.props.listClassName)}
                        data-testid="nodes"
                    >
                        {HeadComponent && (
                            <HeadComponent
                                nodes={this.props.connection.nodes}
                                totalCount={this.props.connection.totalCount}
                                {...this.props.headComponentProps!}
                            />
                        )}
                        {ListComponent === 'table' ? <tbody>{nodes}</tbody> : nodes}
                        {FootComponent && <FootComponent nodes={this.props.connection.nodes} />}
                    </ListComponent>
                )}
                {!this.props.connectionQuery && (
                    <ConnectionNodesSummary
                        summary={summary}
                        displayShowMoreButton={!this.props.loading && !this.props.noShowMore && hasNextPage}
                        showMoreClassName={this.props.showMoreClassName}
                        onShowMore={this.props.onShowMore}
                    />
                )}
            </>
        )
    }
}
