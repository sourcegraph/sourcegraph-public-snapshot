import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs'
import { Activation, ActivationProps } from '../../../shared/src/components/activation/Activation'
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
import {
    fetchAllRepositoriesAndPollIfAnyCloning,
    setAllRepositoriesEnabled,
    setRepositoryEnabled,
    updateAllMirrorRepositories,
    updateMirrorRepository,
} from './backend'

interface RepositoryNodeProps extends ActivationProps {
    node: GQL.IRepository
    onDidUpdate?: () => void
}

interface RepositoryNodeState {
    loading: boolean
    errorDescription?: string
}

class RepositoryNode extends React.PureComponent<RepositoryNodeProps, RepositoryNodeState> {
    public state: RepositoryNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        return (
            <li className="repository-node list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <RepoLink repoName={this.props.node.name} to={this.props.node.url} />
                        {this.props.node.enabled ? (
                            <small
                                data-tooltip="Access to this repository is enabled. All users can view and search it."
                                className="ml-2 text-success"
                            >
                                <CheckIcon className="icon-inline" />
                                Enabled
                            </small>
                        ) : (
                            <small
                                data-tooltip="Access to this repository is disabled. Enable access to it to allow users to view and search it."
                                className="ml-2 text-danger"
                            >
                                <CloseIcon className="icon-inline" />
                                Disabled
                            </small>
                        )}
                        {this.props.node.mirrorInfo.cloneInProgress && (
                            <small className="ml-2 text-success">
                                <LoadingSpinner className="icon-inline" /> Cloning
                            </small>
                        )}
                        {this.props.node.enabled &&
                            !this.props.node.mirrorInfo.cloneInProgress &&
                            !this.props.node.mirrorInfo.cloned && (
                                <small
                                    className="ml-2 text-muted"
                                    data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
                                >
                                    <CloudOutlineIcon className="icon-inline" /> Not yet cloned
                                </small>
                            )}
                    </div>
                    <div className="repository-node__actions">
                        {
                            <Link
                                className="btn btn-secondary btn-sm"
                                to={`/${this.props.node.name}/-/settings`}
                                data-tooltip="Repository settings"
                            >
                                <SettingsIcon className="icon-inline" /> Settings
                            </Link>
                        }{' '}
                        {this.props.node.enabled ? (
                            <button
                                className="btn btn-secondary btn-sm"
                                onClick={this.disableRepository}
                                disabled={this.state.loading}
                                data-tooltip="Disable access to the repository. Users will be unable to view and search it."
                            >
                                Disable
                            </button>
                        ) : (
                            <button
                                className="btn btn-success btn-sm"
                                onClick={this.enableRepository}
                                disabled={this.state.loading}
                                data-tooltip="Enable access to the repository. Users will be able to view and search it."
                            >
                                Enable
                            </button>
                        )}
                    </div>
                </div>
                {this.state.errorDescription && (
                    <div className="alert alert-danger mt-2">{this.state.errorDescription}</div>
                )}
            </li>
        )
    }

    private enableRepository = () => this.setRepositoryEnabled(true)
    private disableRepository = () => this.setRepositoryEnabled(false)

    private setRepositoryEnabled(enabled: boolean): void {
        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        const promises: Promise<any>[] = [setRepositoryEnabled(this.props.node.id, enabled).toPromise()]
        if (enabled) {
            promises.push(updateMirrorRepository({ repository: this.props.node.id }).toPromise())
        }
        Promise.all(promises).then(
            () => {
                if (this.props.onDidUpdate) {
                    this.props.onDidUpdate()
                }
                this.setState({ loading: false })
                activate(this.props.activation)
            },
            err => this.setState({ loading: false, errorDescription: err.message })
        )
    }
}

interface Props extends RouteComponentProps<any>, ActivationProps {}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository, ActivationProps> {}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.PureComponent<Props, {}> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'All',
            id: 'all',
            tooltip: 'Show all repositories',
            args: { enabled: true, disabled: true },
        },
        {
            label: 'Enabled',
            id: 'enabled',
            tooltip: 'Show access-enabled repositories only',
            args: { enabled: true, disabled: false },
        },
        {
            label: 'Disabled',
            id: 'disabled',
            tooltip: 'Show access-disabled repositories only',
            args: { enabled: false, disabled: true },
        },
        {
            label: 'Cloned',
            id: 'cloned',
            tooltip: 'Show cloned repositories only',
            args: { disabled: true, cloned: true, cloneInProgress: false, notCloned: false },
        },
        {
            label: 'Cloning',
            id: 'cloning',
            tooltip: 'Show only repositories that are currently being cloned',
            args: { disabled: true, cloned: false, cloneInProgress: true, notCloned: false },
        },
        {
            label: 'Not cloned',
            id: 'not-cloned',
            tooltip: 'Show only enabled repositories that have not been cloned yet',
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
                {!window.context.sourcegraphDotComMode && (
                    <div className="my-4">
                        <button className="btn btn-secondary" onClick={this.disableAllRepostiories}>
                            Disable all
                        </button>{' '}
                        <button className="btn btn-secondary" onClick={this.enableAllRepostiories}>
                            Enable and clone all
                        </button>
                    </div>
                )}
            </div>
        )
    }

    private queryRepositories = (args: FilteredConnectionQueryArgs) =>
        fetchAllRepositoriesAndPollIfAnyCloning({ ...args })

    private onDidUpdateRepository = () => this.repositoryUpdates.next()

    private enableAllRepostiories = () => this.setAllRepositoriesEnabled(true)
    private disableAllRepostiories = () => this.setAllRepositoriesEnabled(false)

    private setAllRepositoriesEnabled(enabled: boolean): void {
        if (
            enabled &&
            !confirm(
                `Enabling and cloning all repositories may take some time and use significant resources. This will enable and clone all accessible repositories, and is not limited to your current search filter. Enable and clone all repositories?`
            )
        ) {
            return
        }

        eventLogger.log(enabled ? 'EnableAllReposClicked' : 'DisableAllReposClicked')

        const promises: Promise<any>[] = [setAllRepositoriesEnabled(enabled).toPromise()]
        if (enabled) {
            promises.push(updateAllMirrorRepositories().toPromise())
        }
        Promise.all(promises).then(
            () => {
                activate(this.props.activation)
                this.onDidUpdateRepository()
            },
            // If one (or more) repositories fail, still update the UI before re-throwing
            err => {
                this.onDidUpdateRepository()
                throw err
            }
        )
    }
}

function activate(activation?: Activation): void {
    if (activation) {
        activation.update({ EnabledRepository: true })
    }
}
