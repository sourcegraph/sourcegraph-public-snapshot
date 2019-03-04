import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { truncate } from 'lodash'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, map, startWith, switchMap, withLatestFrom } from 'rxjs/operators'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { ThemeProps } from '../../theme'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchSearchResultStats } from '../backend'
import { Sparkline } from './Sparkline'

interface Props extends ThemeProps {
    query: string
    description: string

    actions?: React.ReactNode
    form?: React.ReactNode

    className?: string

    /**
     * The event logged when a user clicks on the link (e.g. 'SavedQueryClick' or 'ExampleSearchClick')
     */
    eventName: string
}

interface State {
    loading: boolean
    approximateResultCount?: string
    sparkline?: number[]
}

export class SavedQueryRow extends React.PureComponent<Props, State> {
    public static defaultProps: Partial<Props> = { className: '' }

    public state: State = { loading: true }

    private componentUpdates = new Subject<Props>()
    private refreshRequested = new Subject<string>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const propsChanges = this.componentUpdates.pipe(startWith(this.props))

        this.subscriptions.add(
            this.refreshRequested
                .pipe(
                    debounceTime(250),
                    withLatestFrom(propsChanges),
                    map(([v]) => v),
                    switchMap(query => fetchSearchResultStats(query)),
                    map(results => ({
                        approximateResultCount: results.approximateResultCount,
                        sparkline: results.sparkline,
                        loading: false,
                    }))
                )
                .subscribe(
                    newState => this.setState(newState as State),
                    err => {
                        this.setState({
                            approximateResultCount: '!',
                            loading: false,
                        })
                        console.error(err)
                    }
                )
        )
        this.refreshRequested.next(this.props.query)
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`saved-query-row ${this.props.className}`}>
                <Link onClick={this.logEvent} to={'/search?' + buildSearchURLQuery(this.props.query)}>
                    <div className="saved-query-row__row">
                        <div className="saved-query-row__row-column">
                            <div className="saved-query-row__description">
                                <span data-tooltip={truncate(this.props.query, { length: 50 })}>
                                    {this.props.description}
                                </span>
                            </div>
                            {this.props.actions}
                        </div>
                        <div className="saved-query-row__results-container">
                            {!this.state.loading && this.state.sparkline && (
                                <div
                                    data-tooltip="Results found in the last 30 days"
                                    className="saved-query-row__sparkline"
                                >
                                    <Sparkline
                                        data={this.state.sparkline}
                                        width={100}
                                        height={40}
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                </div>
                            )}
                            {this.state.loading ? (
                                <div className="saved-query-row__results-items">
                                    <LoadingSpinner className="icon-inline" />
                                </div>
                            ) : (
                                <div className="saved-query-row__results-items">
                                    <div className="saved-query-row__result-count">
                                        {this.state.approximateResultCount}
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </Link>
                {this.props.form}
            </div>
        )
    }

    private logEvent = () => eventLogger.log(this.props.eventName)
}
