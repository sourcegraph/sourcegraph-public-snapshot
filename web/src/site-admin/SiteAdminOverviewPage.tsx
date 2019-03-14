import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import { upperFirst } from 'lodash'
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
import { OverviewItem, OverviewList } from '../components/Overview'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { SiteAdminManagementConsolePassword } from './SiteAdminManagementConsolePassword'
import { UsageChart } from './SiteAdminUsageStatisticsPage'

interface Props extends ActivationProps {
    history: H.History
    overviewComponents: ReadonlyArray<React.ComponentType>
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
            fetchWeeklyActiveUsers().subscribe(stats => this.setState({ stats }), error => this.setState({ error }))
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
            <div className="site-admin-overview-page pt-3">
                <PageTitle title="Overview - Admin" />
                <div className="mb-3">
                    <SiteAdminManagementConsolePassword />
                </div>
                {this.props.overviewComponents.length > 0 && (
                    <div className="mb-4">
                        {this.props.overviewComponents.map((C, i) => (
                            <C key={i} />
                        ))}
                    </div>
                )}
                {!this.state.info && <LoadingSpinner className="icon-inline" />}
                <OverviewList>
                    {this.state.info && (
                        <>
                            {this.props.activation && this.props.activation.completed && (
                                <OverviewItem
                                    title={`${setupPercentage < 100 ? 'Setup Sourcegraph' : 'Status'}`}
                                    defaultExpanded={setupPercentage < 100}
                                    list={true}
                                >
                                    <div>
                                        {this.props.activation.completed && (
                                            <ActivationChecklist
                                                history={this.props.history}
                                                steps={this.props.activation.steps}
                                                completed={this.props.activation.completed}
                                            />
                                        )}
                                    </div>
                                </OverviewItem>
                            )}
                            {this.state.info.repositories !== null && (
                                <OverviewItem link="/explore" actions="Jump to explore page" title="Explore" />
                            )}
                            {this.state.info.repositories !== null && (
                                <OverviewItem
                                    link="/site-admin/repositories"
                                    actions="View all repositories"
                                    title={`${numberWithCommas(this.state.info.repositories)} ${
                                        this.state.info.repositories !== null
                                            ? pluralize('repository', this.state.info.repositories, 'repositories')
                                            : '?'
                                    }`}
                                />
                            )}
                            {this.state.info.users > 1 && (
                                <OverviewItem
                                    link="/site-admin/users"
                                    actions="View or create users"
                                    title={`${numberWithCommas(this.state.info.users)} ${pluralize(
                                        'user',
                                        this.state.info.users
                                    )}`}
                                />
                            )}
                            {this.state.info.orgs > 1 && (
                                <OverviewItem
                                    link="/site-admin/organizations"
                                    actions="View or create organizations"
                                    title={`${numberWithCommas(this.state.info.orgs)} ${pluralize(
                                        'organization',
                                        this.state.info.orgs
                                    )}`}
                                />
                            )}
                            {this.state.info.users > 1 && (
                                <OverviewItem
                                    link="/site-admin/surveys"
                                    actions="View all user surveys"
                                    title={`${numberWithCommas(this.state.info.surveyResponses.totalCount)} ${pluralize(
                                        'user survey response',
                                        this.state.info.surveyResponses.totalCount
                                    )}`}
                                />
                            )}
                            {this.state.info.users > 1 && this.state.stats && (
                                <OverviewItem
                                    title={`${this.state.stats.waus[1].userCount} ${pluralize(
                                        'active user',
                                        this.state.stats.waus[1].userCount
                                    )} last week`}
                                    defaultExpanded={true}
                                >
                                    {this.state.error && (
                                        <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>
                                    )}
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
                                                            className="btn btn-secondary btn-sm"
                                                        >
                                                            <OpenInNewIcon className="icon-inline" /> View all usage
                                                            statistics
                                                        </Link>
                                                    </h3>
                                                </div>
                                            }
                                        />
                                    )}
                                </OverviewItem>
                            )}
                        </>
                    )}
                </OverviewList>
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
