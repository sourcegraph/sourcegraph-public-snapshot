import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import isEqual from 'lodash/isEqual'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { defer } from 'rxjs/observable/defer'
import { catchError } from 'rxjs/operators/catchError'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { retryWhen } from 'rxjs/operators/retryWhen'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { PopoverButton } from '../components/PopoverButton'
import { ChromeExtensionToast, FirefoxExtensionToast } from '../marketing/BrowserExtensionToast'
import { SurveyToast } from '../marketing/SurveyToast'
import { IS_CHROME, IS_FIREFOX } from '../marketing/util'
import { ErrorLike, isErrorLike } from '../util/errors'
import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ECLONEINPROGESS, EREPONOTFOUND, EREVNOTFOUND, ResolvedRev, resolveRev } from './backend'
import { BlobPage } from './blob/BlobPage'
import { DirectoryPage } from './DirectoryPage'
import { FilePathBreadcrumb } from './FilePathBreadcrumb'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevSidebar } from './RepoRevSidebar'
import { RevisionsPopover } from './RevisionsPopover'

interface RepoRevContainerProps extends RouteComponentProps<{ filePath: string }> {
    repo: GQL.IRepository
    rev: string | undefined
    user: GQL.IUser | null
    objectType: 'blob' | 'tree'
    isLightTheme: boolean
}

interface State {
    showSidebar: boolean
    /**
     * The resolved rev or an error if it could not be resolved.
     * `undefined` while loading.
     */
    resolvedRevOrError?: ResolvedRev | ErrorLike
}

/**
 * A container for a repository page that incorporates revisioned Git data. (For example,
 * blob and tree pages are revisioned, but the repository settings page is not.)
 */
export class RepoRevContainer extends React.PureComponent<RepoRevContainerProps, State> {
    public state: State = {
        showSidebar: true,
    }

    private propsUpdates = new Subject<RepoRevContainerProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Fetch repository revision.
        this.subscriptions.add(
            this.propsUpdates
                .pipe(
                    // Pick repoPath and rev out of the props
                    map(props => ({ repoPath: props.repo.uri, rev: props.rev || 'HEAD' })),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    // Reset resolved rev / error state
                    tap(() => this.setState({ resolvedRevOrError: undefined })),
                    switchMap(({ repoPath, rev }) =>
                        defer(() => resolveRev({ repoPath, rev })).pipe(
                            // On a CloneInProgress error, retry after 1s
                            retryWhen(errors =>
                                errors.pipe(
                                    tap(error => {
                                        switch (error.code) {
                                            case ECLONEINPROGESS:
                                                // Display cloning screen to the user and retry
                                                this.setState({ resolvedRevOrError: error })
                                                return
                                            default:
                                                // Display error to the user and do not retry
                                                throw error
                                        }
                                    }),
                                    delay(1000)
                                )
                            ),
                            // Save any error in the sate to display to the user
                            catchError(error => {
                                this.setState({ resolvedRevOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    resolvedRev => {
                        this.setState({ resolvedRevOrError: resolvedRev })
                    },
                    error => {
                        // Should never be reached because errors are caught above
                        console.error(error)
                    }
                )
        )
        this.propsUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: RepoRevContainerProps): void {
        this.propsUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.resolvedRevOrError) {
            // Render nothing while loading
            return null
        }

        if (isErrorLike(this.state.resolvedRevOrError)) {
            // Show error page
            switch (this.state.resolvedRevOrError.code) {
                case ECLONEINPROGESS:
                    return (
                        <HeroPage
                            icon={RepoIcon}
                            title={this.props.repo.uri
                                .split('/')
                                .slice(1)
                                .join('/')}
                            subtitle="Cloning in progress"
                        />
                    )
                case EREPONOTFOUND:
                    return (
                        <HeroPage
                            icon={DirectionalSignIcon}
                            title="404: Not Found"
                            subtitle="The requested repository was not found."
                        />
                    )
                case EREVNOTFOUND:
                    return (
                        <HeroPage
                            icon={DirectionalSignIcon}
                            title="404: Not Found"
                            subtitle="The requested revision was not found."
                        />
                    )
                default:
                    return (
                        <HeroPage
                            icon={ErrorIcon}
                            title="Error"
                            subtitle={upperFirst(this.state.resolvedRevOrError.message)}
                        />
                    )
            }
        }

        return (
            <div className="repo-rev-container">
                {IS_CHROME && <ChromeExtensionToast />}
                {IS_FIREFOX && <FirefoxExtensionToast />}
                <SurveyToast />
                <RepoHeaderActionPortal
                    position="nav"
                    element={
                        <PopoverButton
                            key="repo-rev"
                            className="repo-header__section-btn repo-header__rev"
                            popoverElement={
                                <RevisionsPopover
                                    repo={this.props.repo.id}
                                    repoPath={this.props.repo.uri}
                                    defaultBranch={this.state.resolvedRevOrError.defaultBranch}
                                    currentRev={this.props.rev}
                                    currentCommitID={this.state.resolvedRevOrError.commitID}
                                    history={this.props.history}
                                    location={this.props.location}
                                />
                            }
                            popoverKey="repo-rev"
                            hideOnChange={`${this.props.repo.id}:${this.props.rev || ''}`}
                        >
                            {(this.props.rev && this.props.rev === this.state.resolvedRevOrError.commitID
                                ? this.state.resolvedRevOrError.commitID.slice(0, 7)
                                : this.props.rev) ||
                                this.state.resolvedRevOrError.defaultBranch ||
                                'HEAD'}
                        </PopoverButton>
                    }
                />
                {this.props.match.params.filePath && (
                    <RepoHeaderActionPortal
                        position="nav"
                        element={
                            <FilePathBreadcrumb
                                key="path"
                                repoPath={this.props.repo.uri}
                                rev={this.props.rev}
                                filePath={this.props.match.params.filePath}
                                isDir={this.props.objectType === 'tree'}
                            />
                        }
                    />
                )}
                <RepoHeaderActionPortal
                    position="left"
                    element={<CopyLinkAction key="copy-link" location={this.props.location} />}
                />
                <RepoHeaderActionPortal
                    position="right"
                    element={
                        <GoToPermalinkAction
                            key="go-to-permalink"
                            rev={this.props.rev}
                            commitID={this.state.resolvedRevOrError.commitID}
                            location={this.props.location}
                            history={this.props.history}
                        />
                    }
                />
                <RepoRevSidebar
                    className="repo-rev-container__sidebar"
                    repoID={this.props.repo.id}
                    repoPath={this.props.repo.uri}
                    rev={this.props.rev}
                    commitID={this.state.resolvedRevOrError.commitID}
                    filePath={this.props.match.params.filePath || ''}
                    defaultBranch={this.state.resolvedRevOrError.defaultBranch || 'HEAD'}
                    history={this.props.history}
                    location={this.props.location}
                />
                <div className="repo-rev-container__content">
                    {this.props.objectType === 'tree' && (
                        <DirectoryPage
                            repoPath={this.props.repo.uri}
                            repoDescription={this.props.repo.description}
                            commitID={this.state.resolvedRevOrError.commitID}
                            rev={this.props.rev}
                            filePath={this.props.match.params.filePath || ''}
                            location={this.props.location}
                            history={this.props.history}
                        />
                    )}
                    {this.props.objectType === 'blob' && (
                        <BlobPage
                            repoPath={this.props.repo.uri}
                            commitID={this.state.resolvedRevOrError.commitID}
                            rev={this.props.rev}
                            filePath={this.props.match.params.filePath || ''}
                            location={this.props.location}
                            history={this.props.history}
                            isLightTheme={this.props.isLightTheme}
                        />
                    )}
                </div>
            </div>
        )
    }
}
