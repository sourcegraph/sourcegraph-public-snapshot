import CityIcon from '@sourcegraph/icons/lib/City'
import DocumentIcon from '@sourcegraph/icons/lib/Document'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import UserIcon from '@sourcegraph/icons/lib/User'
import gql from 'graphql-tag'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { queryGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'

interface Props extends RouteComponentProps<any> {}

interface State {
    info?: OverviewInfo
}

/**
 * A page displaying an overview of site admin information.
 */
export class OverviewPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOverview')

        this.subscriptions.add(fetchOverview().subscribe(info => this.setState({ info })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-overview-page">
                <PageTitle title="Overview - Admin" />
                <h2>Site overview</h2>
                <ul className="site-admin-detail-list__list">
                    {this.state.info && (
                        <li className="site-admin-detail-list__item site-admin-overview-page__item">
                            <div className="site-admin-detail-list__header site-admin-overview-page__item-header">
                                <Link
                                    to="/site-admin/repositories"
                                    className="site-admin-overview-page__item-header-link"
                                >
                                    <RepoIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.repositories}{' '}
                                    {pluralize('repository', this.state.info.repositories, 'repositories')}
                                </Link>
                            </div>
                            <div className="site-admin-detail-list__info site-admin-overview-page__item-info">
                                <a href="https://about.sourcegraph.com/docs/server/">
                                    <DocumentIcon className="icon-inline" /> Quickstart: add repositories to search
                                </a>
                                <br />
                                <a href="https://about.sourcegraph.com/docs/server/config/repositories">
                                    <DocumentIcon className="icon-inline" /> Documentation: repositories
                                </a>
                            </div>
                        </li>
                    )}
                    {this.state.info && (
                        <li className="site-admin-detail-list__item site-admin-overview-page__item">
                            <div className="site-admin-detail-list__header site-admin-overview-page__item-header">
                                <Link
                                    to="/site-admin/organizations"
                                    className="site-admin-overview-page__item-header-link"
                                >
                                    <CityIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.orgs} {pluralize('organization', this.state.info.orgs)}
                                </Link>
                            </div>
                            <div className="site-admin-detail-list__info site-admin-overview-page__item-info">
                                <a href="https://about.sourcegraph.com/docs/server/config/organizations">
                                    <DocumentIcon className="icon-inline" /> Documentation: organizations
                                </a>
                            </div>
                        </li>
                    )}
                    {this.state.info && (
                        <li className="site-admin-detail-list__item site-admin-overview-page__item">
                            <div className="site-admin-detail-list__header site-admin-overview-page__item-header">
                                <Link to="/site-admin/users" className="site-admin-overview-page__item-header-link">
                                    <UserIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.users} {pluralize('user', this.state.info.users)}
                                </Link>
                            </div>
                            <div className="site-admin-detail-list__info site-admin-overview-page__item-info">
                                <a href="https://about.sourcegraph.com/docs/server/config/authentication">
                                    <DocumentIcon className="icon-inline" /> Documentation: users
                                </a>
                            </div>
                        </li>
                    )}
                </ul>
            </div>
        )
    }
}

interface OverviewInfo {
    repositories: number
    users: number
    orgs: number
}

function fetchOverview(): Observable<OverviewInfo> {
    return queryGraphQL(gql`
        query Overview {
            repositories {
                totalCount
            }
            users {
                totalCount
            }
            orgs {
                totalCount
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || (!data.repositories || !data.users || !data.orgs)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return {
                repositories: data.repositories.totalCount,
                users: data.users.totalCount,
                orgs: data.orgs.totalCount,
            }
        })
    )
}
