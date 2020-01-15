import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { HeroPage } from '../components/HeroPage'
import { checkMirrorRepositoryConnection } from '../site-admin/backend'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    /** The name of the repository. */
    repo: string

    /** Whether the viewer is a site admin. */
    viewerCanAdminister: boolean
}

interface State {
    /**
     * Whether the option to add the repository should be shown.
     */
    showAdd: boolean

    /**
     * Whether the site admin can add that repository. undefined while loading.
     */
    canAddOrError?: boolean | ErrorLike
}

/**
 * A page informing the user that an error occurred while trying to display the repository. It
 * attempts to present the user with actions to solve the problem.
 */
export class RepositoryNotFoundPage extends React.PureComponent<Props, State> {
    public state: State = {
        showAdd: false,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryError')

        // Show/hide add.
        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (a, b) => a.repo === b.repo && a.viewerCanAdminister === b.viewerCanAdminister
                    ),
                    switchMap(({ repo, viewerCanAdminister }) => {
                        type PartialStateUpdate = Pick<State, 'showAdd' | 'canAddOrError'>
                        if (!viewerCanAdminister) {
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
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    error => console.error(error)
                )
        )

        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Repository not found"
                subtitle={
                    <div className="repository-not-found-page">
                        {that.state.showAdd && (
                            <div className="repository-not-found-page__section mt-3">
                                <div className="repository-not-found-page__section-inner">
                                    <div className="repository-not-found-page__section-description">
                                        {that.state.canAddOrError === undefined && (
                                            <>Checking whether that repository can be added...</>
                                        )}
                                        {(that.state.canAddOrError === false ||
                                            isErrorLike(that.state.canAddOrError)) && (
                                            <>
                                                <p>
                                                    If that is a private repository, check that that site is configured
                                                    with a token that has access to that repository.
                                                </p>

                                                <p>
                                                    If that is a public repository, check that that repository is
                                                    explicitly listed in an{' '}
                                                    <a href="/site-admin/external-services">
                                                        external service configuration
                                                    </a>
                                                    .
                                                </p>
                                            </>
                                        )}
                                        {that.state.canAddOrError === true && (
                                            <>
                                                As a site admin, you can add that repository to Sourcegraph to allow
                                                users to search and view it by{' '}
                                                <a href="/site-admin/external-services">
                                                    connecting an external service
                                                </a>
                                                .
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )}
                        {!that.state.showAdd && <p>To access that repository, contact the Sourcegraph admin.</p>}
                    </div>
                }
            />
        )
    }
}
