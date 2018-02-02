import AddIcon from '@sourcegraph/icons/lib/Add'
import { Checkmark } from '@sourcegraph/icons/lib/Checkmark'
import { Close } from '@sourcegraph/icons/lib/Close'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { catchError } from 'rxjs/operators/catchError'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArgs,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { fetchRepository } from '../repo/backend'
import { RepoLink } from '../repo/RepoLink'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { fetchAllRepositoriesAndPollIfAnyCloning, setRepositoryEnabled, updateMirrorRepository } from './backend'

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
            <li
                key={this.props.node.id}
                className={`site-admin-detail-list__item site-admin-repositories-page__repo site-admin-repositories-page__repo--${
                    this.props.node.enabled ? 'enabled' : 'disabled'
                }`}
            >
                <div className="site-admin-detail-list__header site-admin-repositories-page__repo-header">
                    <RepoLink
                        repoPath={this.props.node.uri}
                        to={`/${this.props.node.uri}/-/settings`}
                        className="site-admin-repositories-page__repo-link"
                    />
                    {this.props.node.enabled ? (
                        <small
                            data-tooltip="Access to this repository is enabled. All users can view and search it."
                            className="site-admin-repositories-page__repo-info site-admin-repositories-page__repo-access"
                        >
                            <Checkmark className="icon-inline" />Enabled
                        </small>
                    ) : (
                        <small
                            data-tooltip="Access to this repository is disabled. Enable access to it to allow users to view and search it."
                            className="site-admin-repositories-page__repo-info site-admin-repositories-page__repo-access"
                        >
                            <Close className="icon-inline" />Disabled
                        </small>
                    )}
                    {this.props.node.mirrorInfo.cloneInProgress && (
                        <small className="site-admin-repositories-page__repo-info">
                            <Loader className="icon-inline" /> Cloning
                        </small>
                    )}
                    {this.props.node.enabled &&
                        !this.props.node.mirrorInfo.cloneInProgress &&
                        !this.props.node.mirrorInfo.cloned && (
                            <small
                                data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
                                className="site-admin-repositories-page__repo-info"
                            >
                                Not yet cloned
                            </small>
                        )}
                </div>
                <div className="site-admin-detail-list__actions site-admin-repositories-page__actions">
                    {
                        <Link
                            className="btn btn-secondary btn-sm site-admin-detail-list__action"
                            to={`/${this.props.node.uri}/-/settings`}
                            data-tooltip="Repository settings"
                        >
                            <GearIcon className="icon-inline" />
                        </Link>
                    }
                    {this.props.node.enabled ? (
                        <button
                            className="btn btn-secondary btn-sm site-admin-detail-list__action"
                            onClick={this.disableRepository}
                            disabled={this.state.loading}
                            data-tooltip="Disable access to the repository. Users will be unable to view and search it."
                        >
                            Disable
                        </button>
                    ) : (
                        <button
                            className="btn btn-success btn-sm site-admin-detail-list__action"
                            onClick={this.enableRepository}
                            disabled={this.state.loading}
                            data-tooltip="Enable access to the repository. Users will be able to view and search it."
                        >
                            Enable
                        </button>
                    )}
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
            },
            err => this.setState({ loading: false, errorDescription: err.message })
        )
    }
}

interface AddPublicRepositoryFormProps {
    onSuccess: () => void
}
interface AddPublicRepositoryFormState {
    repoName: string
    error?: string
    success?: string
}

/**
 * A form for adding a public repository (for now just from https://github.com)
 */
export class AddPublicRepositoryForm extends React.PureComponent<
    AddPublicRepositoryFormProps,
    AddPublicRepositoryFormState
> {
    public state = {
        repoName: '',
        error: undefined,
        success: undefined,
    }

    private inputElement: HTMLInputElement | null = null
    private subscriptions = new Subscription()
    private submits = new Subject<{ repoName: string }>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    mergeMap(({ repoName }) =>
                        fetchRepository({ repoPath: `github.com/${repoName}` }).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ success: undefined, error: error.message })
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    this.setState({ repoName: '', success: 'Repository added', error: undefined })
                    if (this.inputElement) {
                        this.inputElement.focus()
                    }
                    this.props.onSuccess()
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="add-public-repository-form">
                <h3 className="add-public-repository-form__title">Add a public repository</h3>
                {this.state.error && <div className="add-public-repository-form__error">{this.state.error}</div>}
                {this.state.success && <div className="add-public-repository-form__success">{this.state.success}</div>}
                <form className="add-public-repository-form__form" onSubmit={this.onSubmit}>
                    <div className="add-public-repository-form__input-scope">
                        <span className="add-public-repository-form__input-scope-text">{'https://github.com'}/</span>
                    </div>
                    <div className="add-public-repository-form__input-container">
                        <input
                            type="input"
                            className="add-public-repository-form__input-container-input-text form-control"
                            value={this.state.repoName}
                            onChange={this.onPublicRepositoryNameFieldChange}
                            placeholder="organization/repository"
                            ref={input => (this.inputElement = input)}
                        />
                    </div>
                    <button
                        className="add-public-repository-form__add-button btn btn-primary"
                        type="submit"
                        disabled={this.state.repoName === ''}
                    >
                        Add
                    </button>
                </form>
            </div>
        )
    }

    private onPublicRepositoryNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ repoName: e.target.value })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        event.stopPropagation()
        this.submits.next({ repoName: this.state.repoName })
    }
}

interface Props extends RouteComponentProps<any> {}
interface State {
    addPublicRepositoryFormVisible: boolean
}

class FilteredRepositoryConnection extends FilteredConnection<GQL.IRepository> {}

/**
 * A page displaying the repositories on this site.
 */
export class SiteAdminRepositoriesPage extends React.PureComponent<Props, State> {
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
            args: { cloned: true, cloneInProgress: false, notCloned: false },
        },
        {
            label: 'Cloning',
            id: 'cloning',
            tooltip: 'Show repositories that are currently being cloned',
            args: { cloned: false, cloneInProgress: true, notCloned: false },
        },
        {
            label: 'Not cloned',
            id: 'not-cloned',
            tooltip: 'Only show repositories that have not been cloned yet',
            args: { cloned: false, cloneInProgress: false, notCloned: true },
        },
        {
            label: 'Needs index',
            id: 'needs-index',
            tooltip: 'Show repositories that need to be indexed',
            args: { indexed: false },
        },
    ]

    public state: State = {
        addPublicRepositoryFormVisible: false,
    }

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
        const nodeProps: Pick<RepositoryNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateRepository,
        }

        return (
            <div className="site-admin-detail-list site-admin-repositories-page">
                <PageTitle title="Repositories" />
                <div className="site-admin-page__header">
                    <h2 className="site-admin-page__header-title">Repositories</h2>
                    <div className="site-admin-page__actions">
                        <div
                            onClick={this.onDidClickAddPublicRepositoryForm}
                            className="btn btn-primary btn-sm site-admin-page__actions-btn"
                        >
                            <AddIcon className="icon-inline" /> Add public GitHub.com repository
                        </div>
                        <Link
                            to="/site-admin/configuration"
                            className="btn btn-primary btn-sm site-admin-page__actions-btn"
                        >
                            <GearIcon className="icon-inline" /> Configure repositories
                        </Link>
                    </div>
                </div>
                {this.state.addPublicRepositoryFormVisible && (
                    <AddPublicRepositoryForm onSuccess={this.onDidUpdateRepository} />
                )}
                <FilteredRepositoryConnection
                    className="site-admin-page__filtered-connection"
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

    private queryRepositories = (args: FilteredConnectionQueryArgs) =>
        fetchAllRepositoriesAndPollIfAnyCloning({ ...args })

    private onDidUpdateRepository = () => this.repositoryUpdates.next()

    private onDidClickAddPublicRepositoryForm = () => {
        this.setState(state => ({ addPublicRepositoryFormVisible: !state.addPublicRepositoryFormVisible }))
    }
}
