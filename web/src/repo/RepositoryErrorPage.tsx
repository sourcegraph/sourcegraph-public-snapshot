import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import DoNotDisturbIcon from 'mdi-react/DoNotDisturbIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, switchMap, withLatestFrom } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { HeroPage, HeroPageProps } from '../components/HeroPage'
import { checkMirrorRepositoryConnection, setRepositoryEnabled } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    /** The name of the repository. */
    repo: string

    /** The GraphQL ID of the repository, or null if it doesn't exist. */
    repoID: GQL.ID | null

    /** The error that occurred while (unsuccessfully) retrieving the repository, or 'disabled' if
     *  the repository is disabled.
     */
    error: ErrorLike | 'disabled'

    /** Whether the viewer is a site admin. */
    viewerCanAdminister: boolean

    /** Called when the repository is successfully enabled. */
    onDidUpdateRepository?: (update: Partial<GQL.IRepository>) => void
}

interface State {
    /**
     * Whether the option to add the repository should be shown.
     */
    showAdd: boolean

    /**
     * Whether the site admin can add this repository. undefined while loading.
     */
    canAddOrError?: boolean | ErrorLike

    /**
     * Whether the option to enable the repository should be shown.
     */
    showEnable: boolean

    /**
     * Whether the repository was enabled successfully. undefined before being triggered, 'loading' while loading,
     * true if successful, and an error otherwise.
     */
    enabledOrError?: true | 'loading' | ErrorLike
}

/**
 * A page informing the user that an error occurred while trying to display the repository. It
 * attempts to present the user with actions to solve the problem.
 */
export class RepositoryErrorPage extends React.PureComponent<Props, State> {
    public state: State = {
        showAdd: false,
        showEnable: false,
    }

    private componentUpdates = new Subject<Props>()
    private enableClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryError')

        // Show/hide add.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (a, b) =>
                            a.repo === b.repo && a.error === b.error && a.viewerCanAdminister === b.viewerCanAdminister
                    ),
                    switchMap(({ repo, error, viewerCanAdminister }) => {
                        type PartialStateUpdate = Pick<State, 'showAdd' | 'canAddOrError'>
                        if (error === 'disabled' || !viewerCanAdminister) {
                            return of({ showAdd: false, canAddOrError: undefined })
                        }
                        return merge<PartialStateUpdate>(
                            of({ showAdd: true, canAddOrError: undefined }),
                            checkMirrorRepositoryConnection({ name: repo }).pipe(
                                map(c => c.error === null),
                                catchError(error => [asError(error)]),
                                map(c => ({ canAddOrError: c } as PartialStateUpdate))
                            )
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        // Show/hide enable.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (a, b) =>
                            a.repo === b.repo && a.error === b.error && a.viewerCanAdminister === b.viewerCanAdminister
                    ),
                    map(({ error, viewerCanAdminister }) => ({
                        showEnable: error === 'disabled' && viewerCanAdminister,
                    }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        // Handle enable.
        this.subscriptions.add(
            this.enableClicks
                .pipe(
                    withLatestFrom(this.componentUpdates),
                    switchMap(([, { repoID }]) =>
                        merge(
                            of<Pick<State, 'enabledOrError'>>({ enabledOrError: 'loading' }),
                            setRepositoryEnabled(repoID!, true).pipe(
                                map(c => true),

                                // HACK: Delay for gitserver to report the repository as cloning (after
                                // the call to setRepositoryEnabled above, which will trigger a clone).
                                // Without this, there is a race condition where immediately after
                                // clicking this enable button, gitserver reports revision-not-found and
                                // not cloning-in-progress. We need it to report cloning-in-progress so
                                // that the browser polls for the clone to be complete.
                                //
                                // See https://github.com/sourcegraph/sourcegraph/pull/9304.
                                delay(1500),

                                catchError(error => [asError(error)]),
                                map(c => ({ enabledOrError: c } as Pick<State, 'enabledOrError'>))
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => {
                        this.setState(stateUpdate)

                        if (this.props.onDidUpdateRepository && stateUpdate.enabledOrError === true) {
                            this.props.onDidUpdateRepository({ enabled: true })
                        }
                    },
                    error => console.error(error)
                )
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: Props): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let title: string
        let Icon: HeroPageProps['icon']
        if (this.props.error === 'disabled') {
            title = 'Repository disabled'
            Icon = DoNotDisturbIcon
        } else {
            title = 'Repository not found'
            Icon = MapSearchIcon
        }
        return (
            <HeroPage
                icon={Icon}
                title={title}
                subtitle={
                    <div className="repository-error-page">
                        {this.state.showAdd && (
                            <div className="repository-error-page__section mt-3">
                                <div className="repository-error-page__section-inner">
                                    <div className="repository-error-page__section-description">
                                        {this.state.canAddOrError === undefined && (
                                            <>Checking whether this repository can be added...</>
                                        )}
                                        {(this.state.canAddOrError === false ||
                                            isErrorLike(this.state.canAddOrError)) && (
                                            <>
                                                <p>
                                                    The repository can't be added because it is not accessible from any
                                                    code hosts configured on this site.
                                                </p>
                                                <p>
                                                    Check that this site is configured with a token that has access to
                                                    this repository.
                                                </p>
                                            </>
                                        )}
                                        {this.state.canAddOrError === true && (
                                            <>
                                                As a site admin, you can add this repository to Sourcegraph to allow
                                                users to search and view it by{' '}
                                                <a href="/site-admin/external-services">
                                                    connecting an external service
                                                </a>.
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )}
                        {this.state.showEnable && (
                            <div className="repository-error-page__section mt-3">
                                <div className="repository-error-page__section-inner">
                                    <div className="repository-error-page__section-description">
                                        As a site admin, you can enable this repository to allow users to search and
                                        view it.
                                    </div>
                                    <div className="repository-error-page__section-action">
                                        <button
                                            className="btn btn-primary repository-error-page__btn"
                                            onClick={this.enableRepository}
                                            disabled={this.state.enabledOrError === 'loading'}
                                        >
                                            {this.state.enabledOrError === 'loading' ? (
                                                <LoadingSpinner className="icon-inline" />
                                            ) : (
                                                <CheckCircleIcon className="icon-inline" />
                                            )}{' '}
                                            Enable repository
                                        </button>
                                    </div>
                                </div>
                                {isErrorLike(this.state.enabledOrError) && (
                                    <div className="alert alert-danger repository-error-page__alert mt-2">
                                        Error enabling repository: {upperFirst(this.state.enabledOrError.message)}
                                    </div>
                                )}
                            </div>
                        )}
                        {!this.state.showAdd &&
                            !this.state.showEnable && <p>To access this repository, contact the Sourcegraph admin.</p>}
                    </div>
                }
            />
        )
    }

    private enableRepository = () => this.enableClicks.next()
}
