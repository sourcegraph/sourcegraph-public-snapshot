import format from 'date-fns/format'
import React from 'react'
import { Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { BarChart } from '../../components/d3/BarChart'
import { fetchSiteUsageStatistics } from '../../site-admin/backend'
import { ErrorAlert } from '../../components/alerts'

interface Props {
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The site usage statistics, loading, or an error. */
    siteUsageStatisticsOrError: typeof LOADING | GQL.ISiteUsageStatistics | ErrorLike
}

/**
 * An explore section that shows site usage statistics.
 */
export class SiteUsageExploreSection extends React.PureComponent<Props, State> {
    public state: State = { siteUsageStatisticsOrError: LOADING }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            fetchSiteUsageStatistics()
                .pipe(catchError(err => [asError(err)]))
                .subscribe(siteUsageStatisticsOrError => this.setState({ siteUsageStatisticsOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-usage-explore-section">
                <h2>Site usage</h2>
                {isErrorLike(this.state.siteUsageStatisticsOrError) ? (
                    <ErrorAlert error={this.state.siteUsageStatisticsOrError} />
                ) : this.state.siteUsageStatisticsOrError === LOADING ? (
                    <p>Loading...</p>
                ) : (
                    <div className="col-md-10 col-lg-8 mt-4">
                        <BarChart
                            showLabels={true}
                            showLegend={true}
                            width={500}
                            height={200}
                            isLightTheme={this.props.isLightTheme}
                            data={this.state.siteUsageStatisticsOrError.waus.slice(0, 4).map(p => ({
                                xLabel: format(Date.parse(p.startTime) + 1000 * 60 * 60 * 24, 'E, MMM d'),
                                yValues: { 'Weekly users': p.registeredUserCount + p.anonymousUserCount },
                            }))}
                        />
                    </div>
                )}
            </div>
        )
    }
}
