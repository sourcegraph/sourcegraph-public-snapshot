import NoEntryIcon from '@sourcegraph/icons/lib/NoEntry'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { setRepositoryEnabled } from '../site-admin/backend'

interface Props {
    /** The repository that is disabled. */
    repo: GQL.IRepository

    /**
     * Called when the repository is enabled by the site admin.
     */
    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
}

interface State {
    loading: boolean
    error?: Error
}

/**
 * A page informing the user that the repository is disabled. It lets the site admin enable it.
 */
export class RepositoryDisabledPage extends React.PureComponent<Props, State> {
    public state: State = { loading: false }

    private enableClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.enableClicks
                .pipe(
                    tap(() => this.setState({ error: undefined, loading: true })),
                    switchMap(() =>
                        setRepositoryEnabled(this.props.repo.id, true).pipe(
                            catchError(error => {
                                this.setState({ error, loading: false })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    () => {
                        // HACK: Wait (via setTimeout) for gitserver to report the repository as
                        // cloning (after the call to setRepositoryEnabled above, which will trigger
                        // a clone). Without this, there is a race condition where immediately after
                        // clicking this enable button, gitserver reports revision-not-found and not
                        // cloning-in-progress. We need it to report cloning-in-progress so that the
                        // browser polls for the clone to be complete.
                        //
                        // See https://github.com/sourcegraph/sourcegraph/pull/9304.
                        setTimeout(() => {
                            this.setState({ loading: false })
                            this.props.onDidUpdateRepository({ enabled: true })
                        }, 1500)
                    },
                    () => this.setState({ loading: false })
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <HeroPage
                icon={NoEntryIcon}
                title="Repository disabled"
                subtitle={
                    this.props.repo.viewerCanAdminister ? (
                        <div className="repository-disabled-page">
                            <p>As a site admin, you can enable this repository to allow users to search and view it.</p>
                            <button
                                className="btn btn-success repository-disabled-page__btn"
                                onClick={this.enableRepository}
                                disabled={this.state.loading}
                            >
                                Enable repository
                            </button>
                            {this.state.error && (
                                <div className="alert alert-danger repository-disabled-page__alert">
                                    Error enabling repository: <code>{this.state.error.message}</code>
                                </div>
                            )}
                        </div>
                    ) : (
                        'To access this repository, contact the Sourcegraph admin.'
                    )
                }
            />
        )
    }

    private enableRepository = () => this.enableClicks.next()
}
