import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import isEqual from 'lodash/isEqual'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
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
import { ResizablePanel } from '../panel/Panel'
import { ErrorLike, isErrorLike } from '../util/errors'
import { lprToRange, parseHash } from '../util/url'
import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ECLONEINPROGESS, EREPONOTFOUND, EREVNOTFOUND, ResolvedRev, resolveRev } from './backend'
import { BlobPage } from './blob/BlobPage'
import { BlobReferencesPanel } from './blob/references/BlobReferencesPanel'
import { RepositoryCommitsPage } from './commits/RepositoryCommitsPage'
import { DirectoryPage } from './DirectoryPage'
import { FilePathBreadcrumb } from './FilePathBreadcrumb'
import { RepositoryGraphArea } from './graph/RepositoryGraphArea'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevSidebar } from './RepoRevSidebar'
import { EmptyRepositoryPage, RepositoryCloningInProgressPage } from './RepositoryGitDataContainer'
import { RevisionsPopover } from './RevisionsPopover'

interface RepoRevContainerProps extends RouteComponentProps<{}> {
    repo: GQL.IRepository
    rev: string | undefined
    user: GQL.IUser | null
    isLightTheme: boolean
    routePrefix: string
}

interface State {
    showSidebar: boolean
    /**
     * The resolved rev or an error if it could not be resolved.
     * `undefined` while loading.
     */
    resolvedRevOrError?: ResolvedRev | ErrorLike
}

// Run `localStorage.oldBlobPanel=true;location.reload()` to enable the old blob panel in case of issues.
export const useNewBlobPanel = localStorage.getItem('oldBlobPanel') === null || window.context.sourcegraphDotComMode

/** Dev feature flag to make benchmarking the file tree in isolation easier. */
const hideRepoRevContent = localStorage.getItem('hideRepoRevContent')

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
                    return <RepositoryCloningInProgressPage repoName={this.props.repo.uri} />
                case EREPONOTFOUND:
                    return (
                        <HeroPage
                            icon={DirectionalSignIcon}
                            title="404: Not Found"
                            subtitle="The requested repository was not found."
                        />
                    )
                case EREVNOTFOUND:
                    if (!this.props.rev) {
                        return <EmptyRepositoryPage />
                    }

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
                <Switch>
                    {['', '/-/:objectType(blob|tree)/:filePath+'].map(routePath => (
                        <Route
                            path={`${this.props.routePrefix}${routePath}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={routePath === ''}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={(
                                routeComponentProps: RouteComponentProps<{
                                    objectType: 'blob' | 'tree' | undefined
                                    filePath: string | undefined
                                }>
                            ) => {
                                const objectType: 'blob' | 'tree' =
                                    routeComponentProps.match.params.objectType || 'tree'
                                return (
                                    <>
                                        {routeComponentProps.match.params.filePath && (
                                            <RepoHeaderActionPortal
                                                position="nav"
                                                element={
                                                    <FilePathBreadcrumb
                                                        key="path"
                                                        repoPath={this.props.repo.uri}
                                                        rev={this.props.rev}
                                                        filePath={routeComponentProps.match.params.filePath}
                                                        isDir={objectType === 'tree'}
                                                    />
                                                }
                                            />
                                        )}
                                        <RepoRevSidebar
                                            className="repo-rev-container__sidebar"
                                            repoID={this.props.repo.id}
                                            repoPath={this.props.repo.uri}
                                            rev={this.props.rev}
                                            commitID={(this.state.resolvedRevOrError as ResolvedRev).commitID}
                                            filePath={routeComponentProps.match.params.filePath || ''}
                                            isDir={objectType === 'tree'}
                                            defaultBranch={
                                                (this.state.resolvedRevOrError as ResolvedRev).defaultBranch || 'HEAD'
                                            }
                                            history={this.props.history}
                                            location={this.props.location}
                                        />
                                        {!hideRepoRevContent && (
                                            <div className="repo-rev-container__content">
                                                {objectType === 'blob' ? (
                                                    <BlobPage
                                                        repoPath={this.props.repo.uri}
                                                        repoID={this.props.repo.id}
                                                        commitID={
                                                            (this.state.resolvedRevOrError as ResolvedRev).commitID
                                                        }
                                                        rev={this.props.rev}
                                                        filePath={routeComponentProps.match.params.filePath || ''}
                                                        location={this.props.location}
                                                        history={this.props.history}
                                                        isLightTheme={this.props.isLightTheme}
                                                    />
                                                ) : (
                                                    <DirectoryPage
                                                        repoPath={this.props.repo.uri}
                                                        repoID={this.props.repo.id}
                                                        repoDescription={this.props.repo.description}
                                                        commitID={
                                                            (this.state.resolvedRevOrError as ResolvedRev).commitID
                                                        }
                                                        rev={this.props.rev}
                                                        filePath={routeComponentProps.match.params.filePath || ''}
                                                        location={this.props.location}
                                                        history={this.props.history}
                                                        isLightTheme={this.props.isLightTheme}
                                                    />
                                                )}
                                                {useNewBlobPanel ? (
                                                    <ResizablePanel
                                                        isLightTheme={this.props.isLightTheme}
                                                        location={this.props.location}
                                                        history={this.props.history}
                                                    />
                                                ) : (
                                                    this.props.location.hash.includes('$') && (
                                                        <BlobReferencesPanel
                                                            repoPath={this.props.repo.uri}
                                                            commitID={
                                                                (this.state.resolvedRevOrError as ResolvedRev).commitID
                                                            }
                                                            rev={this.props.rev}
                                                            viewState={
                                                                parseHash<'references' | 'references:external'>(
                                                                    this.props.location.hash
                                                                ).viewState
                                                            }
                                                            filePath={routeComponentProps.match.params.filePath || ''}
                                                            position={
                                                                lprToRange(parseHash(this.props.location.hash))!.start
                                                            }
                                                            location={this.props.location}
                                                            history={this.props.history}
                                                            isLightTheme={this.props.isLightTheme}
                                                        />
                                                    )
                                                )}
                                            </div>
                                        )}
                                    </>
                                )
                            }}
                        />
                    ))}
                    <Route
                        path={`${this.props.routePrefix}/-/graph`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        // tslint:disable-next-line:jsx-no-lambda
                        render={(routeComponentProps: RouteComponentProps<{}>) => (
                            <RepositoryGraphArea
                                {...routeComponentProps}
                                repo={this.props.repo}
                                user={this.props.user}
                                rev={this.props.rev}
                                defaultBranch={(this.state.resolvedRevOrError as ResolvedRev).defaultBranch}
                                commitID={(this.state.resolvedRevOrError as ResolvedRev).commitID}
                                routePrefix={this.props.routePrefix}
                            />
                        )}
                    />
                    <Route
                        path={`${this.props.routePrefix}/-/commits`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        // tslint:disable-next-line:jsx-no-lambda
                        render={(routeComponentProps: RouteComponentProps<{}>) => (
                            <RepositoryCommitsPage
                                {...routeComponentProps}
                                repo={this.props.repo}
                                rev={this.props.rev}
                                commitID={(this.state.resolvedRevOrError as ResolvedRev).commitID}
                                location={this.props.location}
                                history={this.props.history}
                            />
                        )}
                    />
                </Switch>
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
            </div>
        )
    }
}
