import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { RepoLink } from '../repo/RepoLink'
import { fetchAllRepositoriesAndPollIfAnyCloning } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'
import { numberWithCommas, pluralize } from '../util/strings'

interface RepositoryNodeProps {
    node: GQL.IRepository
}

const RepositoryNode: React.SFC<RepositoryNodeProps> = ({ node: repo }) => (
    <li key={repo.id} className="list-group-item py-2">
        <RepoLink repoPath={repo.name} to={repo.url} className="explore-page__item-path" />
        {repo.mirrorInfo.cloneInProgress && (
            <small className="ml-2 text-success">
                <LoadingSpinner className="icon-inline" /> Cloning
            </small>
        )}
    </li>
)

interface Props extends RouteComponentProps<any> {
    authenticatedUser: GQL.IUser | null
}

interface State {
    disabledRepositoriesCount?: number | null
}

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
            <div className="explore-page area">
                <div className="area__content">
                    <PageTitle title="Repositories" />
                    <h2>Explore repositories</h2>
                    {this.props.authenticatedUser &&
                        this.props.authenticatedUser.siteAdmin && (
                            <div>
                                <Link to="/site-admin/repositories" className="btn btn-primary">
                                    <SettingsIcon className="icon-inline" /> Configure repositories
                                </Link>
                            </div>
                        )}
                    {this.props.authenticatedUser &&
                        this.props.authenticatedUser.siteAdmin &&
                        typeof this.state.disabledRepositoriesCount === 'number' &&
                        this.state.disabledRepositoriesCount > 0 && (
                            <div className="alert alert-info mt-3 mb-2">
                                {numberWithCommas(this.state.disabledRepositoriesCount)}{' '}
                                {pluralize(
                                    'disabled repository is',
                                    this.state.disabledRepositoriesCount,
                                    'disabled repositories are'
                                )}{' '}
                                not shown here.{' '}
                                <Link to="/site-admin/repositories?filter=disabled">
                                    Enable more repositories in site admin.
                                </Link>
                            </div>
                        )}
                    <FilteredConnection<GQL.IRepository>
                        className="mt-3"
                        listClassName="list-group list-group-flush"
                        noun="repository"
                        pluralNoun="repositories"
                        queryConnection={fetchAllRepositoriesAndPollIfAnyCloning}
                        nodeComponent={RepositoryNode}
                        history={this.props.history}
                        location={this.props.location}
                    />
                </div>
            </div>
        )
    }
}

function fetchDisabledRepositoriesCount(): Observable<number | null> {
    return queryGraphQL(gql`
        query Overview {
            repositories(enabled: false, disabled: true, first: 100) {
                totalCount(precise: true)
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories.totalCount)
    )
}
