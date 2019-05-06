import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { buildSearchURLQuery } from '../../../shared/src/util/url'

export interface SearchUsageStatsProps {
    fetchTopQueries: (count: number) => Observable<GQL.IQueryCount[]>
}

interface SearchUsageStatsState {
    queries: GQL.IQueryCount[]
    count: number
}

const DEFAULT_QUERY_COUNT = 10

export class SearchUsageStats extends React.Component<SearchUsageStatsProps, SearchUsageStatsState> {
    private queryCounts = new Subject<SearchUsageStatsState['count']>()
    private subscriptions = new Subscription()

    constructor(props: SearchUsageStatsProps) {
        super(props)

        this.state = {
            queries: [],
            count: DEFAULT_QUERY_COUNT,
        }

        this.subscriptions.add(
            this.queryCounts
                .pipe(
                    distinctUntilChanged(),
                    switchMap(count => this.props.fetchTopQueries(count).pipe(map(queries => ({ queries, count }))))
                )
                .subscribe(({ queries, count }) => {
                    this.setState({ queries, count })
                })
        )
    }

    public componentDidMount(): void {
        this.queryCounts.next(this.state.count)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <div className="search-usage-stats">
                <h3 className="search-usage-stats__header">
                    <span className="search-usage-stats__header__title">Top search queries</span>
                    <input
                        className="form-control search-usage-stats__header__input"
                        type="number"
                        name="count"
                        value={this.state.count}
                        min={0}
                        onChange={this.onCountChange}
                    />
                </h3>
                <table className="search-usage-stats__top-queries">
                    <thead>
                        <tr>
                            <th>Query</th>
                            <th>Count</th>
                        </tr>
                    </thead>
                    <tbody>
                        {this.state.queries.map(({ query, count }) => (
                            <tr className="search-usage-stats__top-queries__entry">
                                <td className="search-usage-stats__top-queries__entry__cell">
                                    <Link to={`/search?${buildSearchURLQuery(query)}`}>{query}</Link>
                                </td>
                                <td className="search-usage-stats__top-queries__entry__cell">{count}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        )
    }

    private onCountChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        this.queryCounts.next(parseInt(event.target.value, 10))
    }
}
