import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AddIcon from 'mdi-react/AddIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, mergeMap, switchMap } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArgs,
} from '../components/FilteredConnection'
import { Form } from '../components/Form'
import { PageTitle } from '../components/PageTitle'
import { RepoLink } from '../repo/RepoLink'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import {
    addRepository,
    fetchAllRepositoriesAndPollIfAnyCloning,
    setAllRepositoriesEnabled,
    setRepositoryEnabled,
    updateAllMirrorRepositories,
    updateMirrorRepository,
} from './backend'

interface RepositoryNodeProps {
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
                        <RepoLink repoPath={this.props.node.name} to={this.props.node.url} />
                        {this.props.node.enabled ? (
                            <small
                                data-tooltip="Access to this repository is enabled. All users can view and search it."
                                className="ml-2 text-success"
                            >
                                <CheckIcon className="icon-inline" />Enabled
                            </small>
                        ) : (
                            <small
                                data-tooltip="Access to this repository is disabled. Enable access to it to allow users to view and search it."
                                className="ml-2 text-danger"
                            >
                                <CloseIcon className="icon-inline" />Disabled
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
class AddPublicRepositoryForm extends React.PureComponent<AddPublicRepositoryFormProps, AddPublicRepositoryFormState> {
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
                        addRepository(`github.com/${repoName}`).pipe(
                            switchMap(({ id }) => setRepositoryEnabled(id, true)),
                            catchError(error => {
                                console.error(error)
                                eventLogger.log('PublicRepositoryAdditionFailed', {
                                    repositories: { code_host: 'github' },
                                })
                                this.setState({ success: undefined, error: error.message })
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    eventLogger.log('PublicRepositoryAdded', { repositories: { code_host: 'github' } })
                    this.setState({ repoName: '', success: 'Repository added', error: undefined })
                    setTimeout(() => this.setState({ success: undefined }), 1000)
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
                <Form className="add-public-repository-form__form" onSubmit={this.onSubmit}>
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
                            autoFocus={true}
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
                </Form>
                {this.state.error && (
                    <div className="alert alert-danger add-public-repository-form__alert">
                        {upperFirst(this.state.error)}
                    </div>
                )}
                {this.state.success && (
                    <div className="alert alert-success add-public-repository-form__alert">{this.state.success}</div>
                )}
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
            <div className="site-admin-repositories-page">
                <PageTitle title="Repositories - Admin" />
                <h2>Repositories</h2>
                <div>
                    <button
                        onClick={this.toggleAddPublicRepositoryForm}
                        className={`btn ${this.state.addPublicRepositoryFormVisible ? 'btn-secondary' : 'btn-primary'}`}
                    >
                        {this.state.addPublicRepositoryFormVisible ? (
                            <>
                                <CloseIcon className="icon-inline" /> Close
                            </>
                        ) : (
                            <>
                                <AddIcon className="icon-inline" /> Add public GitHub.com repository
                            </>
                        )}
                    </button>{' '}
                    <Link to="/site-admin/configuration" className="btn btn-secondary">
                        <SettingsIcon className="icon-inline" /> Configure repositories
                    </Link>
                </div>
                {this.state.addPublicRepositoryFormVisible && (
                    <AddPublicRepositoryForm onSuccess={this.onDidUpdateRepository} />
                )}
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

    private toggleAddPublicRepositoryForm = () => {
        eventLogger.log('AddPublicRepositoryFormClicked')
        this.setState(state => ({ addPublicRepositoryFormVisible: !state.addPublicRepositoryFormVisible }))
    }

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
            this.onDidUpdateRepository,
            // If one (or more) repositories fail, still update the UI before re-throwing
            err => {
                this.onDidUpdateRepository()
                throw err
            }
        )
    }
}
