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
 * @template C is the GraphQL connection type, such as GQL.IRepositoryConnection.
 * @template N is the node type of the GraphQL connection, such as GQL.IRepository (if C is GQL.IRepositoryConnection)
 * @template NP is the type of the nodeComponent's props
 */
interface Props<C extends Connection<N>, N, NP = {}> {
    history: H.History
    location: H.Location

    /** CSS class name for the root element. */
    className?: string

    /** Called to fetch the connection data to populate this component. */
    queryConnection: (args: FilteredConnectionQueryArgs) => Observable<C>

    /** The component type to use to display each node. */
    nodeComponent: React.ComponentType<{ node: N } & NP>

    /** Props to pass to each nodeComponent. */
    nodeComponentProps?: NP

    /** The English noun (in singular form) describing what this connection contains. */
    noun: string

    /** The English noun (in plural form) describing what this connection contains. */
    pluralNoun: string

    /** An observable that upon emission causes the connection to refresh the data (by calling queryConnection). */
    updates?: Observable<void>

    /** Hides the filter input field. */
    hideFilter?: boolean
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
    nodes: N[]
    totalCount: number
}

/**
 * Displays a collection of items with filtering and pagination. It is called
 * "connection" because it is intended for use with GraphQL, which calls it that
 * (see http://graphql.org/learn/pagination/).
 */
export class FilteredConnection<C extends Connection<N>, N extends GQL.Node> extends React.PureComponent<
    Props<C, N>,
    State<C, N>
> {
    private static DEFAULT_FIRST = 20

    private queryInputChanges = new Subject<string>()
    private showMoreClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public constructor(props: Props<C, N>) {
        super(props)

        const q = new URLSearchParams(this.props.location.search)
        this.state = {
            loading: true,
            query: q.get('q') || '',
            first: parseQueryInt(q, 'first') || FilteredConnection.DEFAULT_FIRST,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.queryInputChanges
                .pipe(
                    startWith(this.state.query),
                    distinctUntilChanged(),
                    tap(query => this.setState({ query })),
                    debounceTime(200),
                    tap(query => {
                        this.props.history.replace({ search: this.urlQuery({ query }) })
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
                        return merge(result, of({ loading: true }).pipe(delay(100), takeUntil(result)))
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
                        this.props.history.replace({ search: this.urlQuery({ first }) })
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
        if (arg.first !== FilteredConnection.DEFAULT_FIRST) {
            q.set('first', String(arg.first))
        }
        return q.toString()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const NodeComponent = this.props.nodeComponent
        return (
            <div className={`filtered-connection ${this.props.className || ''}`}>
                {!this.props.hideFilter && (
                    <form className="filtered-connection__form">
                        <input
                            className="form-control"
                            type="search"
                            placeholder="Search..."
                            name="query"
                            value={this.state.query}
                            onChange={this.onChange}
                        />
                    </form>
                )}
                {!this.state.loading &&
                    this.state.connection &&
                    (this.state.connection.totalCount > 0 ? (
                        <p>
                            <small>
                                <span>
                                    {this.state.connection.totalCount}{' '}
                                    {pluralize(
                                        this.props.noun,
                                        this.state.connection.totalCount,
                                        this.props.pluralNoun
                                    )}{' '}
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
                    ) : (
                        <p>
                            <small>
                                No {this.props.pluralNoun}{' '}
                                {this.state.connectionQuery && (
                                    <span>
                                        matching <strong>{this.state.connectionQuery}</strong>
                                    </span>
                                )}
                            </small>
                        </p>
                    ))}
                {this.state.loading && <Loader className="icon-inline" />}
                {!this.state.loading &&
                    this.state.connection &&
                    this.state.connection.nodes.length > 0 && (
                        <ul className="filtered-connection__nodes">
                            {this.state.connection.nodes.map(node => (
                                <NodeComponent key={node.id} node={node} {...this.props.nodeComponentProps} />
                            ))}
                        </ul>
                    )}
                {!this.state.loading &&
                    this.state.connection &&
                    this.state.connection.nodes.length < this.state.connection.totalCount && (
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
