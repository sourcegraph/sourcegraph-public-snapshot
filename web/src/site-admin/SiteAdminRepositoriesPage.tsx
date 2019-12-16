import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Observable } from 'rxjs'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { RepoLink } from '../../../shared/src/components/RepoLink'
import * as GQL from '../../../shared/src/graphql/schema'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArgs,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { fetchAllRepositoriesAndPollIfEmptyOrAnyCloning } from './backend'
import { ErrorAlert } from '../components/alerts'

interface RepositoryNodeProps extends ActivationProps {
    node: GQL.IRepository
    onDidUpdate?: () => void
}

interface RepositoryNodeState {
    errorDescription?: string
}

class RepositoryNode extends React.PureComponent<RepositoryNodeProps, RepositoryNodeState> {
    public state: RepositoryNodeState = {}

    public render(): JSX.Element | null {
        return (
            <li
                className="repository-node list-group-item py-2"
                data-e2e-repository={this.props.node.name}
                data-e2e-cloned={this.props.node.mirrorInfo.cloned}
            >
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <RepoLink repoName={this.props.node.name} to={this.props.node.url} />
                        {this.props.node.mirrorInfo.cloneInProgress && (
                            <small className="ml-2 text-success">
                                <LoadingSpinner className="icon-inline" /> Cloning
                            </small>
                        )}
                        {!this.props.node.mirrorInfo.cloneInProgress && !this.props.node.mirrorInfo.cloned && (
                            <small
                                className="ml-2 text-muted"
                                data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
                            >
                                <CloudOutlineIcon className="icon-inline" /> Not yet cloned
                            </small>
                        )}
                    </div>
                    <div className="repository-node__actions">
                        {!this.props.node.mirrorInfo.cloneInProgress && !this.props.node.mirrorInfo.cloned && (
                            <Link className="btn btn-sm btn-secondary" to={this.props.node.url}>
                                <CloudDownloadIcon className="icon-inline" /> Clone now
                            </Link>
                        )}{' '}
                        {
                            <Link
                                className="btn btn-secondary btn-sm"
                                to={`/${this.props.node.name}/-/settings`}
                                data-tooltip="Repository settings"
                            >
                                <SettingsIcon className="icon-inline" /> Settings
                            </Link>
                        }{' '}
                    </div>
                </div>
                {this.state.errorDescription && <ErrorAlert className="mt-2" error={this.state.errorDescription} />}
            </li>
        )
    }
}

interface Props extends RouteComponentProps<{}>, ActivationProps {}

class FilteredRepositoryConnection extends FilteredConnection<
    GQL.IRepository,
    Pick<RepositoryNodeProps, 'onDidUpdate'>
> {}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.PureComponent<Props> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'All',
            id: 'all',
            tooltip: 'Show all repositories',
            args: {},
        },
        {
            label: 'Cloned',
            id: 'cloned',
            tooltip: 'Show cloned repositories only',
            args: { cloned: true, cloneInProgress: false, notCloned: false },
        },
        {
            label: 'Cloning',
            id: 'cloning',
            tooltip: 'Show only repositories that are currently being cloned',
            args: { cloned: false, cloneInProgress: true, notCloned: false },
        },
        {
            label: 'Not cloned',
            id: 'not-cloned',
            tooltip: 'Show only repositories that have not been cloned yet',
            args: { cloned: false, cloneInProgress: false, notCloned: true },
        },
        {
            label: 'Needs index',
            id: 'needs-index',
            tooltip: 'Show only repositories that need to be indexed',
            args: { indexed: false },
        },
    ]

    private repositoryUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminRepos')

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
        const nodeProps: Pick<RepositoryNodeProps, 'onDidUpdate' | 'activation'> = {
            onDidUpdate: this.onDidUpdateRepository,
            activation: this.props.activation,
        }

        return (
            <div className="site-admin-repositories-page">
                <PageTitle title="Repositories - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">Repositories</h2>
                </div>
                <p>
                    Repositories are mirrored from connected{' '}
                    <Link to="/site-admin/external-services">external services</Link>.
                </p>
                <FilteredRepositoryConnection
                    className="list-group list-group-flush mt-3"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={this.queryRepositories}
                    nodeComponent={RepositoryNode}
                    nodeComponentProps={nodeProps}
                    updates={this.repositoryUpdates}
                    filters={SiteAdminRepositoriesPage.FILTERS}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryRepositories = (args: FilteredConnectionQueryArgs): Observable<GQL.IRepositoryConnection> =>
        fetchAllRepositoriesAndPollIfEmptyOrAnyCloning({ ...args })

    private onDidUpdateRepository = (): void => this.repositoryUpdates.next()
}
