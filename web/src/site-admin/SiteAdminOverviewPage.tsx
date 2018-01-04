import AddIcon from '@sourcegraph/icons/lib/Add'
import CityIcon from '@sourcegraph/icons/lib/City'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../backend/graphql'
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
export class SiteAdminOverviewPage extends React.Component<Props, State> {
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
            <div className="site-admin-overview-page">
                <PageTitle title="Overview - Admin" />
                {/* <h2>Site overview</h2> */}
                {!this.state.info && <Loader className="icon-inline" />}
                <ul className="site-admin-overview-page__list">
                    {this.state.info && (
                        <li className="site-admin-overview-page__item site-admin-overview-page__item">
                            <div className="site-admin-overview-page__header site-admin-overview-page__item-header">
                                <Link
                                    to="/site-admin/repositories"
                                    className="site-admin-overview-page__item-header-link"
                                >
                                    <RepoIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.repositories}&nbsp;
                                    {pluralize('repository', this.state.info.repositories, 'repositories')}
                                </Link>
                            </div>
                            <div className="site-admin-overview-page__info site-admin-overview-page__item-actions">
                                <Link
                                    to="/site-admin/configuration"
                                    className="btn btn-primary btn-sm site-admin-overview-page__item-action"
                                >
                                    <GearIcon className="icon-inline" /> Configure repositories
                                </Link>
                                <Link
                                    to="/site-admin/repositories"
                                    className="btn btn-secondary btn-sm site-admin-overview-page__item-action"
                                >
                                    View all
                                </Link>
                            </div>
                        </li>
                    )}
                    {this.state.info && (
                        <li className="site-admin-overview-page__item site-admin-overview-page__item">
                            <div className="site-admin-overview-page__header site-admin-overview-page__item-header">
                                <Link to="/site-admin/users" className="site-admin-overview-page__item-header-link">
                                    <UserIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.users}&nbsp;{pluralize('user', this.state.info.users)}
                                </Link>
                            </div>
                            <div className="site-admin-overview-page__info site-admin-overview-page__item-actions">
                                <Link
                                    to="/site-admin/invite-user"
                                    className="btn btn-primary btn-sm site-admin-overview-page__item-action"
                                >
                                    <AddIcon className="icon-inline" /> Invite user
                                </Link>
                                <Link
                                    to="/site-admin/configuration"
                                    className="btn btn-secondary btn-sm site-admin-overview-page__item-action"
                                >
                                    <GearIcon className="icon-inline" /> Configure SSO
                                </Link>
                                <Link
                                    to="/site-admin/users"
                                    className="btn btn-secondary btn-sm site-admin-overview-page__item-action"
                                >
                                    View all
                                </Link>
                            </div>
                        </li>
                    )}
                    {this.state.info && (
                        <li className="site-admin-overview-page__item site-admin-overview-page__item">
                            <div className="site-admin-overview-page__header site-admin-overview-page__item-header">
                                <Link
                                    to="/site-admin/organizations"
                                    className="site-admin-overview-page__item-header-link"
                                >
                                    <CityIcon className="icon-inline site-admin-overview-page__item-header-icon" />
                                    {this.state.info.orgs}&nbsp;{pluralize('organization', this.state.info.orgs)}
                                </Link>
                            </div>
                            <div className="site-admin-overview-page__info site-admin-overview-page__item-actions">
                                <Link
                                    to="/organizations/new"
                                    className="btn btn-primary btn-sm site-admin-overview-page__item-action"
                                >
                                    <AddIcon className="icon-inline" /> Create organization
                                </Link>
                                <Link
                                    to="/site-admin/organizations"
                                    className="btn btn-secondary btn-sm site-admin-overview-page__item-action"
                                >
                                    View all
                                </Link>
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
            site {
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
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.repositories || !data.site.users || !data.site.orgs) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return {
                repositories: data.site.repositories.totalCount,
                users: data.site.users.totalCount,
                orgs: data.site.orgs.totalCount,
            }
        })
    )
}
