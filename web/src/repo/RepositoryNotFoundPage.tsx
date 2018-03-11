import AddIcon from '@sourcegraph/icons/lib/Add'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import Loader from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { switchMap } from 'rxjs/operators/switchMap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { addRepository, checkMirrorRepositoryConnection, setRepositoryEnabled } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../util/errors'

interface Props {
    /** The name of the repository that was not found. */
    repo: string

    /** The not-found error received when the repository was (unsuccessfully) retrieved. */
    notFoundError: ErrorLike

    /** Whether the viewer is authorized to add the repository on this site. */
    viewerCanAddRepository: boolean

    /** Called when the repository is successfully added. */
    onDidAddRepository: () => void
}

interface State {
    /**
     * Whether the site admin can add this repository. undefined while loading.
     */
    canAddOrError?: boolean | ErrorLike

    /**
     * Whether the repository was added successfully. undefined before being triggered, 'loading' while loading,
     * true if successful, and an error otherwise.
     */
    addedOrError?: true | 'loading' | ErrorLike
}

/**
 * A page informing the user that the repository was not found. It lets the site admin add it (if possible) and
 * informs users what they can do if they expected the repository to exist.
 */
export class RepositoryNotFoundPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private addClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryNotFound')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged((a, b) => a.repo === b.repo),
                    switchMap(({ repo }) =>
                        merge<Pick<State, 'canAddOrError'>>(
                            of({ canAddOrError: undefined }),
                            checkMirrorRepositoryConnection({ name: repo }).pipe(
                                map(c => c.error === null),
                                catchError(error => [asError(error)]),
                                map(c => ({ canAddOrError: c } as Pick<State, 'canAddOrError'>)),
                                publishReplay<Pick<State, 'canAddOrError'>>(),
                                refCount()
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

        this.subscriptions.add(
            this.addClicks
                .pipe(
                    withLatestFrom(this.componentUpdates),
                    switchMap(([, { repo }]) =>
                        merge<Pick<State, 'addedOrError'>>(
                            of({ addedOrError: 'loading' }),
                            addRepository(repo).pipe(
                                switchMap(({ id }) => setRepositoryEnabled(id, true)),
                                map(c => true),
                                catchError(error => [asError(error)]),
                                map(c => ({ addedOrError: c } as Pick<State, 'addedOrError'>)),
                                publishReplay<Pick<State, 'addedOrError'>>(),
                                refCount()
                            )
                        )
                    )
                )
                .subscribe(
                    stateUpdate => {
                        this.setState(stateUpdate)
                        if (stateUpdate.addedOrError === true) {
                            this.props.onDidAddRepository()
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
        return (
            <HeroPage
                icon={DirectionalSignIcon}
                title={this.state.canAddOrError === true ? 'Repository not added' : 'Repository not found'}
                subtitle={
                    this.props.viewerCanAddRepository ? (
                        <div className="repository-not-found-page">
                            {this.state.canAddOrError === undefined && (
                                <p>
                                    <Loader className="icon-inline" /> Checking whether this repository can be added...
                                </p>
                            )}
                            {(this.state.canAddOrError === false || isErrorLike(this.state.canAddOrError)) && (
                                <p>
                                    The repository was not found on this Sourcegraph site or any configured code hosts.
                                </p>
                            )}
                            {this.state.canAddOrError === true && (
                                <>
                                    <p>
                                        The repository was not found on Sourcegraph, but it exists on a code host
                                        specified in site configuration.
                                    </p>
                                    <p>
                                        As a site admin, you can add this repository to Sourcegraph to allow users to
                                        search and view it.
                                    </p>
                                    <button
                                        className="btn btn-primary"
                                        onClick={this.addRepository}
                                        disabled={this.state.addedOrError === 'loading'}
                                    >
                                        {this.state.addedOrError === 'loading' ? (
                                            <Loader className="icon-inline" />
                                        ) : (
                                            <AddIcon className="icon-inline" />
                                        )}{' '}
                                        Add repository
                                    </button>
                                    {isErrorLike(this.state.addedOrError) && (
                                        <div className="alert alert-danger repository-not-found-page__alert mt-2">
                                            Error adding repository: <code>{this.state.addedOrError.message}</code>
                                        </div>
                                    )}
                                </>
                            )}
                        </div>
                    ) : (
                        'To access this repository, contact the Sourcegraph admin.'
                    )
                }
            />
        )
    }

    private addRepository = () => this.addClicks.next()
}
