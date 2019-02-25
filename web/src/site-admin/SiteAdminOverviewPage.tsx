import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import { upperFirst } from 'lodash'
import ChartLineIcon from 'mdi-react/ChartLineIcon'
import CityIcon from 'mdi-react/CityIcon'
import EmoticonIcon from 'mdi-react/EmoticonIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PackageVariantIcon from 'mdi-react/PackageVariantIcon'
import RocketIcon from 'mdi-react/RocketIcon'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { ActivationProps, percentageDone } from '../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../shared/src/components/activation/ActivationChecklist'
import { RepositoryIcon } from '../../../shared/src/components/icons' // TODO: Switch to mdi icon
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
    activationCompleted?: { [key: string]: boolean }
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
        if (this.props.activation) {
            this.subscriptions.add(
                this.props.activation.completed.subscribe(completed => {
                    if (completed !== null && completed !== undefined) {
                        this.setState({ activationCompleted: completed })
                    }
                })
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const setupPercentage = percentageDone(this.state.activationCompleted)
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
                            {this.props.activation && this.state.activationCompleted && (
                                <OverviewItem
                                    icon={RocketIcon}
                                    title={`${setupPercentage}% of setup completed`}
                                    defaultExpanded={setupPercentage < 100}
                                >
                                    <div>
                                        <div>
                                            {setupPercentage < 100 ? (
                                                <div>
                                                    <h1>Almost there...</h1>
                                                    <div>
                                                        Complete the steps below to finish onboarding to Sourcegraph.
                                                    </div>
                                                </div>
                                            ) : (
                                                <h2>Setup is complete!</h2>
                                            )}
                                        </div>
                                        {this.props.activation.completed && (
                                            <ActivationChecklist
                                                history={this.props.history}
                                                steps={this.props.activation.steps}
                                                completed={this.state.activationCompleted}
                                            />
                                        )}
                                    </div>
                                </OverviewItem>
                            )}
                            <OverviewItem
                                link="/explore"
                                icon={PackageVariantIcon}
                                actions={
                                    <Link to="/explore" className="btn btn-secondary btn-sm">
                                        <OpenInNewIcon className="icon-inline" /> Explore
                                    </Link>
                                }
                                title="Explore"
                                isBlock={true}
                            />
                            <OverviewItem
                                link="/site-admin/repositories"
                                icon={RepositoryIcon}
                                actions={
                                    <>
                                        <Link to="/site-admin/repositories" className="btn btn-secondary btn-sm">
                                            <OpenInNewIcon className="icon-inline" /> View all
                                        </Link>
                                        <Link to="/site-admin/external-services" className="pr-2">
                                            Configure external services
                                        </Link>
                                    </>
                                }
                                title={`${numberWithCommas(this.state.info.repositories)} ${
                                    this.state.info.repositories !== null
                                        ? pluralize('repository', this.state.info.repositories, 'repositories')
                                        : '?'
                                }`}
                            />
                            <OverviewItem
                                link="/site-admin/users"
                                icon={UserIcon}
                                actions={
                                    <>
                                        <Link to="/site-admin/users" className="btn btn-secondary btn-sm">
                                            <OpenInNewIcon className="icon-inline" /> View all
                                        </Link>
                                        <Link to="/site-admin/users/new" className="pr-2">
                                            Create user account
                                        </Link>
                                    </>
                                }
                                title={`${numberWithCommas(this.state.info.users)} ${pluralize(
                                    'user',
                                    this.state.info.users
                                )}`}
                            />
                            <OverviewItem
                                link="/site-admin/organizations"
                                icon={CityIcon}
                                actions={
                                    <>
                                        <Link to="/site-admin/organizations" className="btn btn-secondary btn-sm">
                                            <OpenInNewIcon className="icon-inline" /> View all
                                        </Link>
                                        <Link to="/organizations/new" className="pr-2">
                                            Create organization
                                        </Link>
                                    </>
                                }
                                title={`${numberWithCommas(this.state.info.orgs)} ${pluralize(
                                    'organization',
                                    this.state.info.orgs
                                )}`}
                            />
                            <OverviewItem
                                link="/site-admin/surveys"
                                icon={EmoticonIcon}
                                actions={
                                    <Link to="/site-admin/surveys" className="btn btn-secondary btn-sm">
                                        <OpenInNewIcon className="icon-inline" /> View all
                                    </Link>
                                }
                                title={`${numberWithCommas(this.state.info.surveyResponses.totalCount)} ${pluralize(
                                    'user survey response',
                                    this.state.info.surveyResponses.totalCount
                                )}`}
                            />
                        </>
                    )}
                    {this.state.stats && (
                        <OverviewItem
                            icon={ChartLineIcon}
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
                                                    <OpenInNewIcon className="icon-inline" /> View all usage statistics
                                                </Link>
                                            </h3>
                                        </div>
                                    }
                                />
                            )}
                        </OverviewItem>
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
