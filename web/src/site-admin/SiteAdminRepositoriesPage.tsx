import GearIcon from '@sourcegraph/icons/lib/Gear'
import Loader from '@sourcegraph/icons/lib/Loader'
import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs/Subject'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { deleteRepository, fetchAllRepositoriesAndPollIfAnyCloning, setRepositoryEnabled } from './backend'

interface RepositoryNodeProps {
    node: GQL.IRepository
    onDidUpdate?: () => void
}

interface RepositoryNodeState {
    loading: boolean
    errorDescription?: string
}

export class RepositoryNode extends React.PureComponent<RepositoryNodeProps, RepositoryNodeState> {
    public state: RepositoryNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        return (
            <li key={this.props.node.id} className="site-admin-detail-list__item site-admin-repositories-page__repo">
                <div className="site-admin-detail-list__header site-admin-repositories-page__repo-header">
                    <Link to={`/${this.props.node.uri}`} className="site-admin-detail-list__name">
                        {this.props.node.uri}
                    </Link>
                    <ul className="site-admin-detail-list__info site-admin-repositories-page__repo-info">
                        {this.props.node.createdAt && (
                            <li>Created: {format(this.props.node.createdAt, 'YYYY-MM-DD')}</li>
                        )}
                        {this.props.node.latest &&
                            this.props.node.latest.cloneInProgress && (
                                <li>
                                    <Loader className="icon-inline" /> Cloning
                                </li>
                            )}
                        <li>
                            Access:{' '}
                            {this.props.node.enabled ? (
                                <span>
                                    <span className="site-admin-repositories-page__repo-enabled">enabled</span> (for all
                                    users)
                                </span>
                            ) : (
                                <span>
                                    <span className="site-admin-repositories-page__repo-disabled">disabled</span> (only
                                    visible to site admins)
                                </span>
                            )}
                        </li>
                    </ul>
                </div>
                <div className="site-admin-detail-list__actions site-admin-repositories-page__actions">
                    {this.props.node.enabled ? (
                        <button
                            className="btn btn-secondary btn-sm"
                            onClick={this.disableRepository}
                            disabled={this.state.loading}
                            title="Disable access to the repository. It will not appear in search results or in the repositories list for non-site-admin users."
                        >
                            Disable access
                        </button>
                    ) : (
                        <button
                            className="btn btn-secondary btn-sm"
                            onClick={this.enableRepository}
                            disabled={this.state.loading}
                        >
                            Enable access
                        </button>
                    )}
                    <button
                        className="btn btn-link btn-sm"
                        onClick={this.deleteRepository}
                        disabled={this.state.loading}
                    >
                        Delete
                    </button>
                    {this.state.errorDescription && (
                        <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                    )}
                </div>
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

        setRepositoryEnabled(this.props.node.id, enabled)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }

    private deleteRepository = () => {
        if (!window.confirm('Really delete this repository? This is irreversible.')) {
            return
        }

        this.setState({
            errorDescription: undefined,
            loading: true,
        })

        deleteRepository(this.props.node.id)
            .toPromise()
            .then(
                () => {
                    this.setState({ loading: false })
                    if (this.props.onDidUpdate) {
                        this.props.onDidUpdate()
                    }
                },
                err => this.setState({ loading: false, errorDescription: err.message })
            )
    }
}

interface Props extends RouteComponentProps<any> {}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.PureComponent<Props> {
    private repositoryUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminRepos')
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<RepositoryNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateRepository,
        }

        return (
            <div className="site-admin-detail-list site-admin-repositories-page">
                <PageTitle title="Repositories" />
                <h2>Repositories</h2>
                <div className="site-admin-page__actions">
                    <Link
                        to="/site-admin/configuration"
                        className="btn btn-primary btn-sm site-admin-page__actions-btn"
                    >
                        <GearIcon className="icon-inline" /> Configure repositories
                    </Link>
                </div>
                <FilteredConnection
                    className="site-admin-page__filtered-connection"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={this.queryRepositories}
                    nodeComponent={RepositoryNode}
                    nodeComponentProps={nodeProps}
                    updates={this.repositoryUpdates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryRepositories = (args: FilteredConnectionQueryArgs) =>
        fetchAllRepositoriesAndPollIfAnyCloning({ ...args, includeDisabled: true })

    private onDidUpdateRepository = () => this.repositoryUpdates.next()
}
