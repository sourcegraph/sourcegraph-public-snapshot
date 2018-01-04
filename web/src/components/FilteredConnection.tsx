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
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { pluralize } from '../util/strings'

interface Props<C extends AbstractConnection<N>, N, NP = {}> {
    history: H.History
    location: H.Location
    className?: string
    queryConnection: (query: string) => Observable<C>
    nodeComponent: React.ComponentType<{ node: N } & NP>
    nodeComponentProps?: NP
    noun: string
    pluralNoun: string
    updates?: Subject<void>
}

interface State<C extends AbstractConnection<N>, N> {
    loading: boolean
    query: string

    connectionQuery?: string
    connection?: C
}

/**
 * See https://facebook.github.io/relay/graphql/connections.htm.
 */
interface AbstractConnection<N> {
    nodes: N[]
    totalCount: number
}

/**
 * Displays a collection of items with filtering and pagination. It is called
 * "connection" because it is intended for use with GraphQL, which calls it that
 * (see http://graphql.org/learn/pagination/).
 */
export class FilteredConnection<C extends AbstractConnection<N>, N extends GQL.Node> extends React.PureComponent<
    Props<C, N>,
    State<C, N>
> {
    private queryInputChanges = new Subject<string>()
    private subscriptions = new Subscription()

    public constructor(props: Props<C, N>) {
        super(props)

        this.state = {
            loading: true,
            query: new URLSearchParams(this.props.location.search).get('q') || '',
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
                        this.props.history.replace({ search: query ? `q=${encodeURIComponent(query)}` : '' })
                    }),
                    switchMap(query => {
                        const result = this.props
                            .queryConnection(query)
                            .pipe(
                                map(
                                    c =>
                                        ({ connection: c, connectionQuery: query, loading: false } as Pick<
                                            State<C, N>,
                                            'connection' | 'loading' | 'connectionQuery'
                                        >)
                                )
                            )
                        return merge(result, of({ loading: true }).pipe(delay(100), takeUntil(result)))
                    })
                )
                .subscribe((stateUpdate: State<C, N>) => this.setState(stateUpdate))
        )

        if (this.props.updates) {
            this.subscriptions.add(
                this.props.updates
                    .pipe(switchMap(() => this.props.queryConnection(this.state.query)))
                    .subscribe((c: C) => this.setState({ connection: c }))
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const NodeComponent = this.props.nodeComponent
        return (
            <div className={`filtered-connection ${this.props.className || ''}`}>
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
                    (this.state.connection.totalCount > 0 ? (
                        <p>
                            <small>
                                {this.state.connection.totalCount}{' '}
                                {pluralize(this.props.noun, this.state.connection.totalCount, this.props.pluralNoun)}{' '}
                                total{' '}
                                {this.state.connectionQuery ? (
                                    <span>
                                        ({this.state.connection.nodes.length}{' '}
                                        {pluralize(
                                            this.props.noun,
                                            this.state.connection.nodes.length,
                                            this.props.pluralNoun
                                        )}{' '}
                                        matching <strong>{this.state.connectionQuery}</strong>)
                                    </span>
                                ) : (
                                    this.state.connection.nodes.length < this.state.connection.totalCount &&
                                    `(showing ${this.state.connectionQuery ? 'matching' : 'first'} ${
                                        this.state.connection.nodes.length
                                    })`
                                )}
                            </small>
                        </p>
                    ) : (
                        <p>
                            No {this.props.pluralNoun}
                            {this.state.connectionQuery ? (
                                <span>
                                    matching <strong>{this.state.connectionQuery}</strong>
                                </span>
                            ) : (
                                ''
                            )}.
                        </p>
                    ))}
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.queryInputChanges.next(e.currentTarget.value)
    }
}
