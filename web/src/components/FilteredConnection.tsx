import Loader from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { combineLatest } from 'rxjs/observable/combineLatest'
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

interface FilterProps {
    /** All filters. */
    filters: FilteredConnectionFilter[]

    /** Called when a filter is selected. */
    onDidSelectFilter: (filter: FilteredConnectionFilter) => void

    /** The ID of the active filter. */
    value: string
}

interface FilterState {}

class FilteredConnectionFilterControl extends React.PureComponent<FilterProps, FilterState> {
    public render(): React.ReactFragment {
        return (
            <div className="filtered-connection-filter-control">
                {this.props.filters.map((filter, i) => (
                    <label key={i} className="filtered-connection-filter-control__item" title={filter.tooltip}>
                        <input
                            className="filtered-connection-filter-control__input"
                            name="filter"
                            type="radio"
                            onChange={this.onChange}
                            value={filter.id}
                            checked={this.props.value === filter.id}
                        />{' '}
                        <div className="filtered-connection-filter-control__label">{filter.label}</div>
                    </label>
                ))}
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        const id = e.currentTarget.value
        const filter = this.props.filters.find(f => f.id === id)!
        this.props.onDidSelectFilter(filter)
    }
}

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

    /**
     * Filters to display next to the filter input field.
     *
     * Filters are mutually exclusive.
     */
    filters?: FilteredConnectionFilter[]
}

/**
 * The arguments for the Props.queryConnection function.
 */
export interface FilteredConnectionQueryArgs {
    first?: number
    query?: string
}

/**
 * A filter to display next to the filter input field.
 */
export interface FilteredConnectionFilter {
    /** The UI label for the filter. */
    label: string

    /**
     * The URL string for this filter (conventionally the label, lowercased and without spaces and punctuation).
     */
    id: string

    /** An optional tooltip to display for this filter. */
    tooltip?: string

    /** Additional query args to pass to the queryConnection function when this filter is enabled. */
    args: { [name: string]: string | number | boolean }
}

interface State<C extends Connection<N>, N> {
    loading: boolean
    query: string
    first: number

    /** The active filter's ID (FilteredConnectionFilter.id), if any. */
    activeFilter: FilteredConnectionFilter | undefined

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
    private activeFilterChanges = new Subject<FilteredConnectionFilter>()
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
            activeFilter: getFilterFromURL(q, this.props.filters),
            first: parseQueryInt(q, 'first') || this.props.defaultFirst!,
        }
    }

    public componentDidMount(): void {
        const activeFilterChanges = this.activeFilterChanges.pipe(startWith(this.state.activeFilter))
        const queryChanges = this.queryInputChanges.pipe(
            distinctUntilChanged(),
            tap(query => this.setState({ query })),
            debounceTime(200),
            startWith(this.state.query)
        )
        const refreshRequests = new Subject<void>()

        this.subscriptions.add(
            this.activeFilterChanges
                .pipe(distinctUntilChanged())
                .subscribe(filter => this.setState({ activeFilter: filter }))
        )

        // Track the last query and filter we used. We only want to show the loader if these change,
        // not when a refresh is requested for the same query/filter (or else there would be jitter).
        let lastQuery: string | undefined
        let lastFilter: FilteredConnectionFilter | undefined
        this.subscriptions.add(
            combineLatest(queryChanges, activeFilterChanges, refreshRequests)
                .pipe(
                    tap(([query, filter]) => {
                        if (!this.props.noUpdateURLQuery) {
                            this.props.history.replace({ search: this.urlQuery({ query, filter }) })
                        }
                    }),
                    switchMap(([query, filter]) => {
                        const result = this.props
                            .queryConnection({ first: this.state.first, query, ...(filter ? filter.args : {}) })
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

                        const showLoading = query !== lastQuery || filter !== lastFilter
                        lastQuery = query
                        lastFilter = filter
                        return showLoading
                            ? merge(result, of({ loading: true }).pipe(delay(250), takeUntil(result)))
                            : result
                    })
                )
                .subscribe((stateUpdate: State<C, N>) => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.showMoreClicks
                .pipe(map(() => this.state.first * 2))
                .subscribe(first => this.setState({ first }, () => refreshRequests.next()))
        )

        if (this.props.updates) {
            this.subscriptions.add(this.props.updates.subscribe(c => refreshRequests.next()))
        }

        // Reload collection when the query callback changes.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ queryConnection }) => queryConnection),
                    distinctUntilChanged(),
                    tap(() => this.focusFilter())
                )
                .subscribe(() => this.setState({ loading: true }, () => refreshRequests.next()))
        )
        this.componentUpdates.next(this.props)
    }

    private urlQuery(arg: { first?: number; query?: string; filter?: FilteredConnectionFilter }): string {
        if (!arg.first) {
            arg.first = this.state.first
        }
        if (!arg.query) {
            arg.query = this.state.query
        }
        if (!arg.filter) {
            arg.filter = this.state.activeFilter
        }
        const q = new URLSearchParams()
        if (arg.query) {
            q.set('q', arg.query)
        }
        if (arg.first !== this.props.defaultFirst) {
            q.set('first', String(arg.first))
        }
        if (arg.filter && this.props.filters && arg.filter !== this.props.filters[0]) {
            q.set('filter', arg.filter.id)
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
                        {this.props.filters &&
                            this.state.activeFilter && (
                                <FilteredConnectionFilterControl
                                    filters={this.props.filters}
                                    onDidSelectFilter={this.onDidSelectFilter}
                                    value={this.state.activeFilter.id}
                                />
                            )}
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

    private onDidSelectFilter = (filter: FilteredConnectionFilter) => this.activeFilterChanges.next(filter)

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

function getFilterFromURL(
    q: URLSearchParams,
    filters: FilteredConnectionFilter[] | undefined
): FilteredConnectionFilter | undefined {
    if (filters === undefined || filters.length === 0) {
        return undefined
    }
    const id = q.get('filter')
    if (id !== null) {
        const filter = filters.find(f => f.id === id)
        if (filter) {
            return filter
        }
    }
    return filters[0] // default
}
