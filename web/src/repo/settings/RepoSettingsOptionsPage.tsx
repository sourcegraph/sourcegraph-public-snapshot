import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { REPO_DELETE_CONFIRMATION_MESSAGE } from '.'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { deleteRepository, setRepositoryEnabled } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepository } from './backend'
import { ActionContainer } from './components/ActionContainer'

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
}

interface State {
    /**
     * The repository object, refreshed after we make changes that modify it.
     */
    repo: GQL.IRepository

    loading: boolean
    error?: string
}

/**
 * The repository settings options page.
 */
export class RepoSettingsOptionsPage extends React.PureComponent<Props, State> {
    private repoUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettings')

        this.subscriptions.add(
            this.repoUpdates
                .pipe(switchMap(() => fetchRepository(this.props.repo.name)))
                .subscribe(repo => this.setState({ repo }), err => this.setState({ error: err.message }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-settings-options-page">
                <PageTitle title="Repository settings" />
                <h2>Settings</h2>
                {this.state.loading && <LoadingSpinner className="icon-inline" />}
                {this.state.error && <div className="alert alert-danger">{upperFirst(this.state.error)}</div>}
                <Form>
                    <div className="form-group">
                        <label htmlFor="repo-settings-options-page__name">Repository name</label>
                        <input
                            id="repo-settings-options-page__name"
                            type="text"
                            className="form-control"
                            readOnly={true}
                            disabled={true}
                            value={this.state.repo.name}
                            required={true}
                            spellCheck={false}
                            autoCapitalize="off"
                            autoCorrect="off"
                            aria-describedby="repo-settings-options-page__name-help"
                        />
                    </div>
                </Form>
                <ActionContainer
                    title={this.state.repo.enabled ? 'Disable access' : 'Enable access'}
                    description={
                        this.state.repo.enabled
                            ? 'Disable access to the repository to prevent users from searching and browsing the repository.'
                            : 'The repository is disabled. Enable it to allow users to search and view the repository.'
                    }
                    buttonClassName={this.state.repo.enabled ? 'btn-danger' : 'btn-success'}
                    buttonLabel={this.state.repo.enabled ? 'Disable access' : 'Enable access'}
                    flashText="Updated"
                    run={this.state.repo.enabled ? this.disableRepository : this.enableRepository}
                />
                <ActionContainer
                    title="Delete repository"
                    description="Permanently removes this repository and all associated data from Sourcegraph. The original repository on the code host is not affected. If this repository was added by a configured code host, then it will be re-added during the next sync."
                    buttonClassName="btn-danger"
                    buttonLabel="Delete this repository"
                    run={this.deleteRepository}
                />
            </div>
        )
    }

    private enableRepository = () =>
        setRepositoryEnabled(this.state.repo.id, true)
            .toPromise()
            .then(() => {
                this.repoUpdates.next()
                this.props.onDidUpdateRepository({ enabled: true })
            })
    private disableRepository = () =>
        setRepositoryEnabled(this.state.repo.id, false)
            .toPromise()
            .then(() => {
                this.repoUpdates.next()
                this.props.onDidUpdateRepository({ enabled: false })
            })

    private deleteRepository = () => {
        if (!window.confirm(REPO_DELETE_CONFIRMATION_MESSAGE)) {
            return Promise.resolve()
        }

        return deleteRepository(this.state.repo.id)
            .toPromise()
            .then(() => {
                this.repoUpdates.next()
                this.props.history.push('/site-admin/repositories')
            })
    }
}
