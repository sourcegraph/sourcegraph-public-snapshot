import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { ActivationProps, percentageDone } from '../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../shared/src/components/activation/ActivationChecklist'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { numberWithCommas, pluralize } from '../../../shared/src/util/strings'
import { queryGraphQL } from '../backend/graphql'
import { Collapsible } from '../components/Collapsible'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { UsageChart } from './SiteAdminUsageStatisticsPage'
import { ErrorAlert } from '../components/alerts'

interface Props extends ActivationProps {
    history: H.History
    overviewComponents: readonly React.ComponentType[]
    isLightTheme: boolean
}

interface State {
    info?: OverviewInfo
    stats?: GQL.ISiteUsageStatistics
    error?: Error
}

const fetchOverview: () => Observable<OverviewInfo> = () =>
    queryGraphQL(gql`
        query Overview {
            repositories {
                totalCount(precise: true)
            }
            users {
                totalCount
            }
            organizations {
                totalCount
            }
            surveyResponses {
                totalCount
                averageScore
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => ({
            repositories: data.repositories.totalCount,
            users: data.users.totalCount,
            orgs: data.organizations.totalCount,
            surveyResponses: data.surveyResponses,
        }))
    )

const fetchWeeklyActiveUsers: () => Observable<GQL.ISiteUsageStatistics> = () =>
    queryGraphQL(gql`
        query WAUs {
            site {
                usageStatistics {
                    waus {
                        userCount
                        registeredUserCount
                        anonymousUserCount
                        startTime
                    }
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.usageStatistics)
    )

/**
 * A page displaying an overview of site admin information.
 */
export class SiteAdminOverviewPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOverview')

        this.subscriptions.add(fetchOverview().subscribe(info => this.setState({ info })))
        this.subscriptions.add(
            fetchWeeklyActiveUsers().subscribe(
                stats => this.setState({ stats }),
                error => this.setState({ error })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let setupPercentage = 0
        if (this.props.activation) {
            setupPercentage = percentageDone(this.props.activation.completed)
        }
        return (
            <div className="site-admin-overview-page py-3">
                <PageTitle title="Overview - Admin" />
                {this.props.overviewComponents.length > 0 && (
                    <div className="mb-4">
                        {this.props.overviewComponents.map((C, i) => (
                            <C key={i} />
                        ))}
                    </div>
                )}
                {!this.state.info && <LoadingSpinner className="icon-inline" />}
                <div className="list-group">
                    {this.state.info && (
                        <>
                            {this.props.activation && this.props.activation.completed && (
                                <Collapsible
                                    title={<>{setupPercentage < 100 ? 'Set up Sourcegraph' : 'Status'}</>}
                                    defaultExpanded={setupPercentage < 100}
                                    className="list-group-item"
                                    titleClassName="h5 mb-0 font-weight-normal p-2"
                                >
                                    {this.props.activation.completed && (
                                        <ActivationChecklist
                                            history={this.props.history}
                                            steps={this.props.activation.steps}
                                            completed={this.props.activation.completed}
                                            className="border-top"
                                        />
                                    )}
                                </Collapsible>
                            )}
                            {this.state.info.repositories !== null && (
                                <Link
                                    to="/site-admin/repositories"
                                    className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                                >
                                    {numberWithCommas(this.state.info.repositories)}{' '}
                                    {pluralize('repository', this.state.info.repositories, 'repositories')}
                                </Link>
                            )}
                            {this.state.info.users > 1 && (
                                <Link
                                    to="/site-admin/users"
                                    className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                                >
                                    {numberWithCommas(this.state.info.users)} {pluralize('user', this.state.info.users)}
                                </Link>
                            )}
                            {this.state.info.orgs > 1 && (
                                <Link
                                    to="/site-admin/organizations"
                                    className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                                >
                                    {numberWithCommas(this.state.info.orgs)}{' '}
                                    {pluralize('organization', this.state.info.orgs)}
                                </Link>
                            )}
                            {this.state.info.users > 1 && (
                                <Link
                                    to="/site-admin/surveys"
                                    className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                                >
                                    {numberWithCommas(this.state.info.surveyResponses.totalCount)}{' '}
                                    {pluralize('user survey response', this.state.info.surveyResponses.totalCount)}
                                </Link>
                            )}
                            {this.state.info.users > 1 && this.state.stats && (
                                <Collapsible
                                    title={
                                        <>
                                            {this.state.stats.waus[1].userCount}{' '}
                                            {pluralize('active user', this.state.stats.waus[1].userCount)} last week
                                        </>
                                    }
                                    defaultExpanded={true}
                                    className="list-group-item"
                                    titleClassName="h5 mb-0 font-weight-normal p-2"
                                >
                                    {this.state.error && <ErrorAlert className="mb-3" error={this.state.error} />}
                                    {this.state.stats && (
                                        <UsageChart
                                            {...this.props}
                                            stats={this.state.stats}
                                            chartID="waus"
                                            showLegend={false}
                                            header={
                                                <div className="site-admin-overview-page__detail-header">
                                                    <h2>Weekly unique users</h2>
                                                    <h3>
                                                        <Link
                                                            to="/site-admin/usage-statistics"
                                                            className="btn btn-secondary"
                                                        >
                                                            View all usage statistics{' '}
                                                            <OpenInNewIcon className="icon-inline" />
                                                        </Link>
                                                    </h3>
                                                </div>
                                            }
                                        />
                                    )}
                                </Collapsible>
                            )}
                        </>
                    )}
                </div>
            </div>
        )
    }
}

interface OverviewInfo {
    repositories: number | null
    users: number
    orgs: number
    surveyResponses: {
        totalCount: number
        averageScore: number
    }
}
