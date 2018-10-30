import format from 'date-fns/format'
import React from 'react'
import { Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { BarChart } from '../../components/d3/BarChart'
import { fetchSiteAnalytics } from '../../site-admin/backend'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'

interface Props {
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The site activity, loading, or an error. */
    siteActivityOrError: typeof LOADING | GQL.ISiteActivity | ErrorLike
}

/**
 * An explore section that shows site usage.
 */
export class SiteUsageExploreSection extends React.PureComponent<Props, State> {
    public state: State = { siteActivityOrError: LOADING }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            fetchSiteAnalytics()
                .pipe(catchError(err => [asError(err)]))
                .subscribe(siteActivityOrError => this.setState({ siteActivityOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="extensions-explore-section">
                <h2>Site usage</h2>
                {isErrorLike(this.state.siteActivityOrError) ? (
                    <div className="alert alert-danger">Error: {this.state.siteActivityOrError.message}</div>
                ) : this.state.siteActivityOrError === LOADING ? (
                    <p>Loading...</p>
                ) : (
                    <div className="col-md-10 col-lg-8 mt-4">
                        <BarChart
                            showLabels={true}
                            showLegend={true}
                            width={500}
                            height={200}
                            isLightTheme={this.props.isLightTheme}
                            data={this.state.siteActivityOrError.waus.slice(0, 4).map(p => ({
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
