import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { isEqual, upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { defer, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, retryWhen, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { PopoverButton } from '../components/PopoverButton'
import { ChromeExtensionToast, FirefoxExtensionToast } from '../marketing/BrowserExtensionToast'
import { SurveyToast } from '../marketing/SurveyToast'
import { IS_CHROME, IS_FIREFOX } from '../marketing/util'
import { ResizablePanel } from '../panel/Panel'
import { getModeFromPath } from '../util'
import { ErrorLike, isErrorLike } from '../util/errors'
import { CodeIntelStatusIndicator } from './actions/CodeIntelStatusIndicator'
import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { CloneInProgressError, ECLONEINPROGESS, EREPONOTFOUND, EREVNOTFOUND, ResolvedRev, resolveRev } from './backend'
import { BlobPage } from './blob/BlobPage'
import { RepositoryCommitsPage } from './commits/RepositoryCommitsPage'
import { FilePathBreadcrumb } from './FilePathBreadcrumb'
import { RepositoryGraphArea } from './graph/RepositoryGraphArea'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevSidebar } from './RepoRevSidebar'
import { EmptyRepositoryPage, RepositoryCloningInProgressPage } from './RepositoryGitDataContainer'
import { RevisionsPopover } from './RevisionsPopover'
import { TreePage } from './TreePage'

interface RepoRevContainerProps extends RouteComponentProps<{}>, ExtensionsProps {
    repo: GQL.IRepository
    rev: string
    user: GQL.IUser | null
    isLightTheme: boolean
    onHelpPopoverToggle: () => void
    routePrefix: string

    /**
     * The resolved rev or an error if it could not be resolved. This value lives in RepoContainer (this
     * component's parent) but originates from this component.
     */
    resolvedRevOrError?: ResolvedRev | ErrorLike

    /** Called when the resolvedRevOrError state in this component's parent should be updated. */
    onResolvedRevOrError: (v: ResolvedRev | ErrorLike | undefined) => void
}

interface State {
    showSidebar: boolean
}

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
                    map(props => ({ repoPath: props.repo.uri, rev: props.rev })),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    // Reset resolved rev / error state
                    tap(() => this.props.onResolvedRevOrError(undefined)),
                    switchMap(({ repoPath, rev }) =>
                        defer(() => resolveRev({ repoPath, rev })).pipe(
                            // On a CloneInProgress error, retry after 1s
                            retryWhen(errors =>
                                errors.pipe(
                                    tap(error => {
                                        switch (error.code) {
                                            case ECLONEINPROGESS:
                                                // Display cloning screen to the user and retry
                                                this.props.onResolvedRevOrError(error)
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
                                this.props.onResolvedRevOrError(error)
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    resolvedRev => {
                        this.props.onResolvedRevOrError(resolvedRev)
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
        if (!this.props.resolvedRevOrError) {
            // Render nothing while loading
            return null
        }

        if (isErrorLike(this.props.resolvedRevOrError)) {
            // Show error page
            switch (this.props.resolvedRevOrError.code) {
                case ECLONEINPROGESS:
                    return (
                        <RepositoryCloningInProgressPage
                            repoName={this.props.repo.uri}
                            progress={(this.props.resolvedRevOrError as CloneInProgressError).progress}
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
                            subtitle={upperFirst(this.props.resolvedRevOrError.message)}
                        />
                    )
            }
        }

        const resolvedRev = this.props.resolvedRevOrError
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
                                    defaultBranch={this.props.resolvedRevOrError.defaultBranch}
                                    currentRev={this.props.rev}
                                    currentCommitID={this.props.resolvedRevOrError.commitID}
                                    history={this.props.history}
                                    location={this.props.location}
                                />
                            }
                            popoverKey="repo-rev"
                            hideOnChange={`${this.props.repo.id}:${this.props.rev || ''}`}
                        >
                            {(this.props.rev && this.props.rev === this.props.resolvedRevOrError.commitID
                                ? this.props.resolvedRevOrError.commitID.slice(0, 7)
                                : this.props.rev) ||
                                this.props.resolvedRevOrError.defaultBranch ||
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
                                const filePath = routeComponentProps.match.params.filePath || '' // empty string is root
                                const mode = getModeFromPath(filePath)
                                return (
                                    <>
                                        {filePath && (
                                            <>
                                                <RepoHeaderActionPortal
                                                    position="nav"
                                                    element={
                                                        <FilePathBreadcrumb
                                                            key="path"
                                                            repoPath={this.props.repo.uri}
                                                            rev={this.props.rev}
                                                            filePath={filePath}
                                                            isDir={objectType === 'tree'}
                                                        />
                                                    }
                                                />
                                                {objectType === 'blob' && (
                                                    <RepoHeaderActionPortal
                                                        position="right"
                                                        priority={-10}
                                                        element={
                                                            <CodeIntelStatusIndicator
                                                                key="code-intel-status"
                                                                userIsSiteAdmin={
                                                                    !!this.props.user && this.props.user.siteAdmin
                                                                }
                                                                repoPath={this.props.repo.uri}
                                                                rev={this.props.rev}
                                                                commitID={resolvedRev.commitID}
                                                                filePath={filePath}
                                                                mode={mode}
                                                            />
                                                        }
                                                    />
                                                )}
                                            </>
                                        )}
                                        <RepoRevSidebar
                                            className="repo-rev-container__sidebar"
                                            repoID={this.props.repo.id}
                                            repoPath={this.props.repo.uri}
                                            rev={this.props.rev}
                                            commitID={(this.props.resolvedRevOrError as ResolvedRev).commitID}
                                            filePath={routeComponentProps.match.params.filePath || ''}
                                            isDir={objectType === 'tree'}
                                            defaultBranch={
                                                (this.props.resolvedRevOrError as ResolvedRev).defaultBranch || 'HEAD'
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
                                                            (this.props.resolvedRevOrError as ResolvedRev).commitID
                                                        }
                                                        rev={this.props.rev}
                                                        filePath={routeComponentProps.match.params.filePath || ''}
                                                        mode={mode}
                                                        extensions={this.props.extensions}
                                                        location={this.props.location}
                                                        history={this.props.history}
                                                        isLightTheme={this.props.isLightTheme}
                                                    />
                                                ) : (
                                                    <TreePage
                                                        repoPath={this.props.repo.uri}
                                                        repoID={this.props.repo.id}
                                                        repoDescription={this.props.repo.description}
                                                        commitID={
                                                            (this.props.resolvedRevOrError as ResolvedRev).commitID
                                                        }
                                                        rev={this.props.rev}
                                                        filePath={routeComponentProps.match.params.filePath || ''}
                                                        location={this.props.location}
                                                        history={this.props.history}
                                                        isLightTheme={this.props.isLightTheme}
                                                        onHelpPopoverToggle={this.props.onHelpPopoverToggle}
                                                    />
                                                )}
                                                <ResizablePanel
                                                    isLightTheme={this.props.isLightTheme}
                                                    location={this.props.location}
                                                    history={this.props.history}
                                                />
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
                                defaultBranch={(this.props.resolvedRevOrError as ResolvedRev).defaultBranch}
                                commitID={(this.props.resolvedRevOrError as ResolvedRev).commitID}
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
                                commitID={(this.props.resolvedRevOrError as ResolvedRev).commitID}
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
                            commitID={this.props.resolvedRevOrError.commitID}
                            location={this.props.location}
                            history={this.props.history}
                        />
                    }
                />
            </div>
        )
    }
}
