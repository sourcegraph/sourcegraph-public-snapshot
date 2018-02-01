import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../backend/graphql'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { RepoFileLink } from '../components/RepoFileLink'
import { fetchAllRepositoriesAndPollIfAnyCloning } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'
import { pluralize } from '../util/strings'

interface RepositoryNodeProps {
    node: GQL.IRepository
}

export const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node: repo }) => (
    <li key={repo.id} className="repo-browser__item">
        <div className="repo-browser__item-header">
            <Link to={`/${repo.uri}`} className="repo-browser__item-path">
                <RepoFileLink repoPath={repo.uri} disableLinks={true} />
            </Link>
            {repo.mirrorInfo.cloneInProgress && (
                <span className="repo-browser__item-cloning">
                    <small>
                        <Loader className="icon-inline" /> Cloning
                    </small>
                </span>
            )}
        </div>
        <div className="repo-browser__item-spacer" />
        <div className="repo-browser__item-actions">
            {repo.viewerCanAdminister && (
                <Link
                    to={`/${repo.uri}/-/settings`}
                    className="btn btn-secondary btn-sm repo-browser__item-action"
                    data-tooltip="Repository settings"
                >
                    <GearIcon className="icon-inline" />
                </Link>
            )}
            <Link
                to={`/${repo.uri}`}
                className="btn btn-secondary btn-sm repo-browser__item-action"
                data-tooltip="Search and explore this repository"
            >
                <RepoIcon className="icon-inline" />&nbsp;View
            </Link>
        </div>
    </li>
)

interface RepoBrowserProps extends RouteComponentProps<any> {
    user: GQL.IUser | null
}

interface State {
    disabledRepositoriesCount?: number | null
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

export class RepoBrowser extends React.PureComponent<RepoBrowserProps, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('Browse')
        this.subscriptions.add(
            fetchDisabledRepositoriesCount().subscribe(disabledRepositoriesCount =>
                this.setState({ disabledRepositoriesCount })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-browser">
                <PageTitle title="Repositories" />
                <div className="repo-browser__header">
                    <h2>Repositories</h2>
                    {this.props.user &&
                        this.props.user.siteAdmin && (
                            <div className="repo-browser__actions">
                                <Link
                                    to="/site-admin/repositories"
                                    className="btn btn-primary btn-sm site-admin-page__actions-btn"
                                >
                                    <GearIcon className="icon-inline" /> Repositories (site admin)
                                </Link>
                            </div>
                        )}
                </div>
                {this.props.user &&
                    this.props.user.siteAdmin &&
                    this.state.disabledRepositoriesCount && (
                        <div className="alert alert-notice repo-browser__notice">
                            {this.state.disabledRepositoriesCount}{' '}
                            {pluralize(
                                'disabled repository is',
                                this.state.disabledRepositoriesCount,
                                'disabled repositories are'
                            )}{' '}
                            not shown here.{' '}
                            <Link
                                data-tooltip="Enabling a repository makes it accessible and searchable to all users."
                                to="/site-admin/repositories?filter=disabled"
                            >
                                Enable repositories
                            </Link>
                        </div>
                    )}
                <FilteredRepositoryConnection
                    className="repo-browser__filtered-connection"
                    listClassName="repo-browser__items"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={fetchAllRepositoriesAndPollIfAnyCloning}
                    nodeComponent={RepositoryNode}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}

function fetchDisabledRepositoriesCount(): Observable<number | null> {
    return queryGraphQL(gql`
        query Overview {
            site {
                repositories(enabled: false, disabled: true, first: 100) {
                    totalCount(precise: true)
                    pageInfo {
                        hasNextPage
                    }
                }
                users {
                    totalCount
                }
                orgs {
                    totalCount
                }
                hasCodeIntelligence
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.repositories) {
                throw createAggregateError(errors)
            }
            return data.site.repositories.totalCount
        })
    )
}
