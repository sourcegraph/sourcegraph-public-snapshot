import Loader from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { pluralize } from '../util/strings'

/**
 * Props for the FilteredConnection component.
 *
 * @template C The GraphQL connection type, such as GQL.IRepositoryConnection.
 * @template N The node type of the GraphQL connection, such as GQL.IRepository (if C is GQL.IRepositoryConnection)
 * @template NP Props passed to `nodeComponent` in addition to `{ node: N }`
 */
interface FilteredConnectionProps<C extends Connection<N>, N extends GQL.Node, NP = {}> {
    history: H.History
    location: H.Location

    /** CSS class name for the root element. */
    className?: string

    /** Whether to display it more compactly. */
    compact?: boolean

    /** CSS class name for the list element (<ul>). */
    listClassName?: string

    /** Called to fetch the connection data to populate this component. */
    queryConnection: (args: FilteredConnectionQueryArgs) => Observable<C>

    /** The component type to use to display each node. */
    nodeComponent: React.ComponentType<{ node: N } & NP>

    /** Props to pass to each nodeComponent in addition to `{ node: N }`. */
    nodeComponentProps?: NP

    /** The English noun (in singular form) describing what this connection contains. */
    noun: string

    /** The English noun (in plural form) describing what this connection contains. */
    pluralNoun: string

    /** An observable that upon emission causes the connection to refresh the data (by calling queryConnection). */
    updates?: Observable<void>

    /** The number of items to fetch, by default. */
    defaultFirst?: number

    /** Hides the filter input field. */
    hideFilter?: boolean

    /** Autofocuses the filter input field. */
    autoFocus?: boolean

    /** Do not update the URL query string to reflect the filter and pagination state. */
    noUpdateURLQuery?: boolean

    /** Do not show a "Show more" button. */
    noShowMore?: boolean

    /** Do not show a count summary if all nodes are visible in the list's first page. */
    noSummaryIfAllNodesVisible?: boolean
}

/**
 * The arguments for the Props.queryConnection function.
 */
export interface FilteredConnectionQueryArgs {
    first?: number
    query?: string
}

interface State<C extends Connection<N>, N> {
    loading: boolean
    query: string
    first: number

    connectionQuery?: string
    connection?: C
}

/**
 * See https://facebook.github.io/relay/graphql/connections.htm.
 */
interface Connection<N> {
    /**
     * The list of items (nodes) in this connection's current page.
     */
    nodes: N[]

    /**
     * The total count of items in the connection (not subject to pagination). The type accounts
     * for all known GraphQL XyzConnection types.
     *
     * If the value is a number, then the precise total count is known. If null, then the total
     * count was not precisely computable for this particular query (but might be for other queries).
     * If undefined, then the resolver never supports producing a total count.
     *
     * In the future, the UI might show `null` differently from `undefined`, but for now, the
     * distinction is maintained solely for typechecking to pass.
     */
    totalCount?: number | null

    /**
     * If set, indicates whether there is a next page. Not all GraphQL XyzConnection types return
     * pageInfo (if not, then they generally all do return totalCount).
     */
    pageInfo?: { hasNextPage: boolean }
}

/**
 * Displays a collection of items with filtering and pagination. It is called
 * "connection" because it is intended for use with GraphQL, which calls it that
 * (see http://graphql.org/learn/pagination/).
 *
 * @template C The GraphQL connection type, such as GQL.IRepositoryConnection.
 * @template N The node type of the GraphQL connection, such as GQL.IRepository (if C is GQL.IRepositoryConnection)
 * @template NP Props passed to `nodeComponent` in addition to `{ node: N }`
 */
export class FilteredConnection<
    N extends GQL.Node,
    NP = {},
    C extends Connection<N> = Connection<N>
> extends React.PureComponent<FilteredConnectionProps<C, N, NP>, State<C, N>> {
    public static defaultProps: Partial<FilteredConnectionProps<any, any>> = {
        defaultFirst: 20,
    }

    private queryInputChanges = new Subject<string>()
    private showMoreClicks = new Subject<void>()
    private componentUpdates = new Subject<FilteredConnectionProps<C, N, NP>>()
    private subscriptions = new Subscription()

    private filterRef: HTMLInputElement | null = null

    public constructor(props: FilteredConnectionProps<C, N, NP>) {
        super(props)

        const q = new URLSearchParams(this.props.location.search)
        this.state = {
            loading: true,
            query: q.get('q') || '',
            first: parseQueryInt(q, 'first') || this.props.defaultFirst!,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.queryInputChanges
                .pipe(
                    distinctUntilChanged(),
                    tap(query => this.setState({ query })),
                    debounceTime(200),
                    startWith(this.state.query),
                    tap(query => {
                        if (!this.props.noUpdateURLQuery) {
                            this.props.history.replace({ search: this.urlQuery({ query }) })
                        }
                    }),
                    switchMap(query => {
                        const result = this.props
                            .queryConnection({ first: this.state.first, query })
                            .pipe(
                                map(
                                    c =>
                                        ({ connection: c, connectionQuery: query, loading: false } as Pick<
                                            State<C, N>,
                                            'connection' | 'loading' | 'connectionQuery'
                                        >)
                                ),
                                publishReplay(),
                                refCount()
                            )
                        return merge(result, of({ loading: true }).pipe(delay(250), takeUntil(result)))
                    })
                )
                .subscribe((stateUpdate: State<C, N>) => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.showMoreClicks
                .pipe(
                    map(() => this.state.first * 2),
                    tap(first => {
                        this.setState({ first })
                        if (!this.props.noUpdateURLQuery) {
                            this.props.history.replace({ search: this.urlQuery({ first }) })
                        }
                    }),
                    switchMap(first => this.props.queryConnection({ first, query: this.state.query }))
                )
                .subscribe(c => this.setState({ connection: c }))
        )

        if (this.props.updates) {
            this.subscriptions.add(
                this.props.updates
                    .pipe(
                        switchMap(() =>
                            this.props.queryConnection({ first: this.state.first, query: this.state.query })
                        )
                    )
                    .subscribe(c => this.setState({ connection: c }))
            )
        }

        // Reload collection when the query callback changes.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ queryConnection }) => queryConnection),
                    distinctUntilChanged(),
                    tap(() => this.focusFilter()),
                    switchMap(queryConnection => queryConnection({ first: this.state.first, query: this.state.query }))
                )
                .subscribe(c => this.setState({ connection: c }))
        )
    }

    private urlQuery(arg: { first?: number; query?: string }): string {
        if (!arg.first) {
            arg.first = this.state.first
        }
        if (!arg.query) {
            arg.query = this.state.query
        }
        const q = new URLSearchParams()
        if (arg.query) {
            q.set('q', arg.query)
        }
        if (arg.first !== this.props.defaultFirst) {
            q.set('first', String(arg.first))
        }
        return q.toString()
    }

    public componentWillReceiveProps(nextProps: FilteredConnectionProps<C, N, NP>): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const NodeComponent = this.props.nodeComponent

        const hasNextPage =
            this.state.connection &&
            ((this.state.connection.pageInfo && this.state.connection.pageInfo.hasNextPage) ||
                (typeof this.state.connection.totalCount === 'number' &&
                    this.state.connection.nodes.length < this.state.connection.totalCount))

        let summary: React.ReactFragment | undefined
        if (
            !this.state.loading &&
            this.state.connection &&
            (!this.props.noSummaryIfAllNodesVisible || this.state.connection.nodes.length === 0 || hasNextPage)
        ) {
            if (typeof this.state.connection.totalCount === 'number' && this.state.connection.totalCount > 0) {
                summary = (
                    <p className="filtered-connection__summary">
                        <small>
                            <span>
                                {this.state.connection.totalCount}{' '}
                                {pluralize(this.props.noun, this.state.connection.totalCount, this.props.pluralNoun)}{' '}
                                {this.state.connectionQuery ? (
                                    <span>
                                        {' '}
                                        matching <strong>{this.state.connectionQuery}</strong>
                                    </span>
                                ) : (
                                    'total'
                                )}
                            </span>{' '}
                            {this.state.connection.nodes.length < this.state.connection.totalCount &&
                                `(showing first ${this.state.connection.nodes.length})`}
                        </small>
                    </p>
                )
            } else if (this.state.connection.pageInfo && this.state.connection.pageInfo.hasNextPage) {
                // No total count to show, but it will show a 'Show more' button.
            } else {
                summary = (
                    <p className="filtered-connection__summary">
                        <small>
                            No {this.props.pluralNoun}{' '}
                            {this.state.connectionQuery && (
                                <span>
                                    matching <strong>{this.state.connectionQuery}</strong>
                                </span>
                            )}
                        </small>
                    </p>
                )
            }
        }

        const compactnessClass = `filtered-connection--${this.props.compact ? 'compact' : 'noncompact'}`
        return (
            <div className={`filtered-connection ${compactnessClass} ${this.props.className || ''}`}>
                {!this.props.hideFilter && (
                    <form className="filtered-connection__form">
                        <input
                            className="form-control filtered-connection__filter"
                            type="search"
                            placeholder={`Search ${this.props.pluralNoun}...`}
                            name="query"
                            value={this.state.query}
                            onChange={this.onChange}
                            autoFocus={this.props.autoFocus}
                            autoComplete="off"
                            autoCorrect="off"
                            autoCapitalize="off"
                            ref={this.setFilterRef}
                            spellCheck={false}
                        />
                    </form>
                )}
                {this.state.loading && <Loader className="icon-inline filtered-connection__loader" />}
                {this.state.connectionQuery && summary}
                {!this.state.loading &&
                    this.state.connection &&
                    this.state.connection.nodes.length > 0 && (
                        <ul className={`filtered-connection__nodes ${this.props.listClassName || ''}`}>
                            {this.state.connection.nodes.map(node => (
                                <NodeComponent key={node.id} node={node} {...this.props.nodeComponentProps} />
                            ))}
                        </ul>
                    )}
                {!this.state.connectionQuery && summary}
                {!this.state.loading &&
                    !this.props.noShowMore &&
                    this.state.connection &&
                    hasNextPage && (
                        <button
                            className="btn btn-secondary btn-sm filtered-connection__show-more"
                            onClick={this.onClickShowMore}
                        >
                            Show more
                        </button>
                    )}
            </div>
        )
    }

    private setFilterRef = (e: HTMLInputElement | null) => {
        this.filterRef = e
        if (e && this.props.autoFocus) {
            // TODO(sqs): The 30 msec delay is needed, or else the input is not
            // reliably focused. Find out why.
            setTimeout(() => e.focus(), 30)
        }
    }

    private focusFilter = () => {
        if (this.filterRef) {
            this.filterRef.focus()
        }
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.queryInputChanges.next(e.currentTarget.value)
    }

    private onClickShowMore: React.MouseEventHandler<HTMLButtonElement> = e => {
        this.showMoreClicks.next()
    }
}

function parseQueryInt(q: URLSearchParams, name: string): number | null {
    const s = q.get(name)
    if (s === null) {
        return null
    }
    const n = parseInt(s, 10)
    if (n > 0) {
        return n
    }
    return null
}
