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
import { RepoLink } from '../repo/RepoLink'
import { fetchAllRepositoriesAndPollIfAnyCloning } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'
import { pluralize } from '../util/strings'

interface RepositoryNodeProps {
    node: GQL.IRepository
}

export const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node: repo }) => (
    <li key={repo.id} className="explore-page__item">
        <div className="explore-page__item-header">
            <RepoLink repoPath={repo.uri} className="explore-page__item-path" />
            {repo.mirrorInfo.cloneInProgress && (
                <span className="explore-page__item-cloning">
                    <small>
                        <Loader className="icon-inline" /> Cloning
                    </small>
                </span>
            )}
        </div>
        <div className="explore-page__item-spacer" />
        <div className="explore-page__item-actions">
            {repo.viewerCanAdminister && (
                <Link
                    to={`/${repo.uri}/-/settings`}
                    className="btn btn-secondary btn-sm explore-page__item-action"
                    data-tooltip="Repository settings"
                >
                    <GearIcon className="icon-inline" />
                </Link>
            )}
            <Link
                to={`/${repo.uri}`}
                className="btn btn-secondary btn-sm explore-page__item-action"
                data-tooltip="Search and explore this repository"
            >
                <RepoIcon className="icon-inline" />&nbsp;View
            </Link>
        </div>
    </li>
)

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser | null
}

interface State {
    disabledRepositoriesCount?: number | null
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

/**
 * A page for exploring the repositories on this site.
 */
export class ExplorePage extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('Explore')
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
            <div className="explore-page">
                <PageTitle title="Repositories" />
                <div className="explore-page__header">
                    <h2>Explore repositories</h2>
                    {this.props.user &&
                        this.props.user.siteAdmin && (
                            <div className="explore-page__actions">
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
                        <div className="alert alert-notice explore-page__notice">
                            {this.state.disabledRepositoriesCount}{' '}
                            {pluralize(
                                'disabled repository is',
                                this.state.disabledRepositoriesCount,
                                'disabled repositories are'
                            )}{' '}
                            not shown here.{' '}
                            <Link to="/site-admin/repositories?filter=disabled">
                                Enable repositories in site admin.
                            </Link>
                        </div>
                    )}
                <FilteredRepositoryConnection
                    className="explore-page__filtered-connection"
                    listClassName="explore-page__items"
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
