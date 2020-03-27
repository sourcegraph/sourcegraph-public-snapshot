import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Observable } from 'rxjs'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { RepoLink } from '../../../shared/src/components/RepoLink'
import * as GQL from '../../../shared/src/graphql/schema'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { fetchAllRepositoriesPermissions } from './backend'
import { ErrorAlert } from '../components/alerts'
import { Timestamp } from '../components/time/Timestamp'

interface RepositoryPermissionsNodeProps extends ActivationProps {
    node: GQL.IRepositoryPermissions
    onDidUpdate?: () => void
}

interface RepositoryPermissionsNodeState {
    errorDescription?: string
}

class RepositoryPermissionsNode extends React.PureComponent<
    RepositoryPermissionsNodeProps,
    RepositoryPermissionsNodeState
> {
    public state: RepositoryPermissionsNodeState = {}

    public render(): JSX.Element | null {
        return (
            <li className="repository-node list-group-item py-2" data-e2e-repository={this.props.node.repository.name}>
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <RepoLink repoName={this.props.node.repository.name} to={this.props.node.repository.url} />
                    </div>
                    <div className="repository-node__actions">
                        <small className="ml-2 text-muted">
                            <Timestamp date={this.props.node.permissions.updatedAt} />
                        </small>{' '}
                    </div>
                </div>
                {this.state.errorDescription && <ErrorAlert className="mt-2" error={this.state.errorDescription} />}
            </li>
        )
    }
}

interface Props extends RouteComponentProps<{}>, ActivationProps {}

class FilteredRepositoryPermissionsConnection extends FilteredConnection<
    GQL.IRepositoryPermissions,
    Pick<RepositoryPermissionsNodeProps, 'onDidUpdate'>
> {}

/**
 * A page displaying the repositories permissions on this site.
 */
export class SiteAdminRepositoriesPermissionsPage extends React.PureComponent<Props> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'Most recent synced',
            id: 'most-recent-synced',
            tooltip: 'Sort by most recent synced',
            args: {},
        },
        {
            label: 'Least recent synced',
            id: 'least-recent-synced',
            tooltip: 'Sort by least recent synced',
            args: { descending: false },
        },
    ]

    private repositoryUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminReposPerms')

        // Refresh global alert about enabling repositories when the user visits here.
        refreshSiteFlags()
            .toPromise()
            .then(null, err => console.error(err))
    }

    public componentWillUnmount(): void {
        // Remove global alert about enabling repositories when the user navigates away from here.
        refreshSiteFlags()
            .toPromise()
            .then(null, err => console.error(err))
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<RepositoryPermissionsNodeProps, 'onDidUpdate' | 'activation'> = {
            onDidUpdate: this.onDidUpdateRepository,
            activation: this.props.activation,
        }

        return (
            <div className="site-admin-repositories-page">
                <PageTitle title="Repositories permissions - Admin" />
                <h2>Repositories permissions</h2>
                <p>
                    Repositories permissions are synced from connected{' '}
                    <Link to="/site-admin/external-services">code host connections</Link>.
                </p>
                <FilteredRepositoryPermissionsConnection
                    className="list-group list-group-flush mt-3"
                    noun="entry"
                    pluralNoun="entries"
                    queryConnection={this.queryRepositoriesPermissions}
                    nodeComponent={RepositoryPermissionsNode}
                    nodeComponentProps={nodeProps}
                    updates={this.repositoryUpdates}
                    filters={SiteAdminRepositoriesPermissionsPage.FILTERS}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryRepositoriesPermissions = (
        args: FilteredConnectionQueryArgs
    ): Observable<GQL.IRepositoryPermissionsConnection> => fetchAllRepositoriesPermissions({ ...args })

    private onDidUpdateRepository = (): void => this.repositoryUpdates.next()
}
