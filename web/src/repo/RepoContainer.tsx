import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap, tap } from 'rxjs/operators'
import { redirectToExternalHost } from '.'
import { WorkspaceRootWithMetadata } from '../../../shared/src/api/client/model'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { makeRepoURI } from '../../../shared/src/util/url'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { searchQueryForRepoRev } from '../search'
import { queryUpdates } from '../search/input/QueryInput'
import { parseBrowserRepoURL, ParsedRepoRev, parseRepoRev } from '../util/url'
import { GoToCodeHostAction } from './actions/GoToCodeHostAction'
import { EREPONOTFOUND, EREPOSEEOTHER, fetchRepository, RepoSeeOtherError, ResolvedRev } from './backend'
import { RepositoryBranchesArea } from './branches/RepositoryBranchesArea'
import { RepositoryCommitPage } from './commit/RepositoryCommitPage'
import { RepositoryCompareArea } from './compare/RepositoryCompareArea'
import { RepositoryReleasesArea } from './releases/RepositoryReleasesArea'
import { RepoHeader, RepoHeaderActionButton, RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RepoRevContainer, RepoRevContainerRoute } from './RepoRevContainer'
import { RepositoryErrorPage } from './RepositoryErrorPage'
import { RepositoryGitDataContainer } from './RepositoryGitDataContainer'
import { RepoSettingsArea } from './settings/RepoSettingsArea'
import { RepositoryStatsArea } from './stats/RepositoryStatsArea'

const RepoPageNotFound: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="The repository page was not found." />
)

export interface RepoContainerProps
    extends RouteComponentProps<{ repoRevAndRest: string }>,
        SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps {
    repoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute>
    repoHeaderActionButtons: ReadonlyArray<RepoHeaderActionButton>
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface RepoRevContainerState extends ParsedRepoRev {
    filePath?: string
    rest?: string

    /**
     * The fetched repository or an error if occurred.
     * `undefined` while loading.
     */
    repoOrError?: GQL.IRepository | ErrorLike

    /**
     * The resolved rev or an error if it could not be resolved. `undefined` while loading. This value comes from
     * this component's child RepoRevContainer, but it lives here because it's used by other children than just
     * RepoRevContainer.
     */
    resolvedRevOrError?: ResolvedRev | ErrorLike

    /** The external links to show in the repository header, if any. */
    externalLinks?: GQL.IExternalLink[]

    repoHeaderContributionsLifecycleProps?: RepoHeaderContributionsLifecycleProps
}

/**
 * Renders a horizontal bar and content for a repository page.
 */
export class RepoContainer extends React.Component<RepoContainerProps, RepoRevContainerState> {
    private routeMatchChanges = new Subject<{ repoRevAndRest: string }>()
    private repositoryUpdates = new Subject<Partial<GQL.IRepository>>()
    private revResolves = new Subject<ResolvedRev | ErrorLike | undefined>()
    private subscriptions = new Subscription()

    constructor(props: RepoContainerProps) {
        super(props)

        this.state = {
            ...parseURLPath(props.match.params.repoRevAndRest),
        }
    }

    public componentDidMount(): void {
        const parsedRouteChanges = this.routeMatchChanges.pipe(
            map(({ repoRevAndRest }) => parseURLPath(repoRevAndRest))
        )

        // Fetch repository.
        const repositoryChanges = parsedRouteChanges.pipe(
            map(({ repoName }) => repoName),
            distinctUntilChanged()
        )
        this.subscriptions.add(
            repositoryChanges
                .pipe(
                    tap(() => this.setState({ repoOrError: undefined })),
                    switchMap(repoName =>
                        fetchRepository({ repoName }).pipe(
                            catchError(error => {
                                switch (error.code) {
                                    case EREPOSEEOTHER:
                                        redirectToExternalHost((error as RepoSeeOtherError).redirectURL)
                                        return []
                                }
                                this.setState({ repoOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    repo => {
                        this.setState({ repoOrError: repo })
                    },
                    err => {
                        console.error(err)
                    }
                )
        )

        // Update resolved revision in state
        this.revResolves.subscribe(resolvedRevOrError => this.setState({ resolvedRevOrError }))

        // Update header and other global state.
        this.subscriptions.add(
            parsedRouteChanges.subscribe(({ repoName, rev, rawRev, rest }) => {
                this.setState({ repoName, rev, rawRev, rest })

                queryUpdates.next(searchQueryForRepoRev(repoName, rev))
            })
        )

        this.routeMatchChanges.next(this.props.match.params)

        // Merge in repository updates.
        this.subscriptions.add(
            this.repositoryUpdates.subscribe(update =>
                this.setState(({ repoOrError }) => ({ repoOrError: { ...repoOrError, ...update } as GQL.IRepository }))
            )
        )

        // Update the Sourcegraph extensions model to reflect the current workspace root.
        this.subscriptions.add(
            this.revResolves
                .pipe(
                    map(resolvedRevOrError => {
                        let roots: WorkspaceRootWithMetadata[] | null = null
                        if (resolvedRevOrError && !isErrorLike(resolvedRevOrError)) {
                            roots = [
                                {
                                    uri: makeRepoURI({
                                        repoName: this.state.repoName,
                                        rev: resolvedRevOrError.commitID,
                                    }),
                                    inputRevision: this.state.rev || '',
                                },
                            ]
                        }
                        this.props.extensionsController.services.model.model.next({
                            ...this.props.extensionsController.services.model.model.value,
                            roots,
                        })
                    })
                )
                .subscribe()
        )
        // Clear the Sourcegraph extensions model's roots when navigating away.
        this.subscriptions.add(() =>
            this.props.extensionsController.services.model.model.next({
                ...this.props.extensionsController.services.model.model.value,
                roots: null,
            })
        )
    }

    public componentWillReceiveProps(props: RepoContainerProps): void {
        if (props.match.params !== this.props.match.params) {
            this.routeMatchChanges.next(props.match.params)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.repoOrError) {
            // Render nothing while loading
            return null
        }

        const { repoName, filePath, commitRange, position, range } = parseBrowserRepoURL(
            location.pathname + location.search + location.hash
        )
        const viewerCanAdminister = !!this.props.authenticatedUser && this.props.authenticatedUser.siteAdmin

        if (isErrorLike(this.state.repoOrError)) {
            // Display error page
            switch (this.state.repoOrError.code) {
                case EREPONOTFOUND:
                    return (
                        <RepositoryErrorPage
                            repo={repoName}
                            repoID={null}
                            error={this.state.repoOrError}
                            viewerCanAdminister={viewerCanAdminister}
                        />
                    )
                default:
                    return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={this.state.repoOrError.message} />
            }
        }

        const repoMatchURL = `/${this.state.repoOrError.name}`

        const transferProps = {
            repo: this.state.repoOrError,
            authenticatedUser: this.props.authenticatedUser,
            isLightTheme: this.props.isLightTheme,
            repoMatchURL,
            settingsCascade: this.props.settingsCascade,
            platformContext: this.props.platformContext,
            extensionsController: this.props.extensionsController,
            ...this.state.repoHeaderContributionsLifecycleProps,
        }

        const isSettingsPage =
            location.pathname === `${repoMatchURL}/-/settings` ||
            location.pathname.startsWith(`${repoMatchURL}/-/settings/`)

        return (
            <div className="repo-container w-100 d-flex flex-column">
                <RepoHeader
                    actionButtons={this.props.repoHeaderActionButtons}
                    rev={this.state.rev}
                    repo={this.state.repoOrError}
                    resolvedRev={this.state.resolvedRevOrError}
                    platformContext={this.props.platformContext}
                    extensionsController={this.props.extensionsController}
                    onLifecyclePropsChange={this.onRepoHeaderContributionsLifecyclePropsChange}
                    location={this.props.location}
                    history={this.props.history}
                />
                <RepoHeaderContributionPortal
                    position="right"
                    key="go-to-code-host"
                    priority={2}
                    element={
                        <GoToCodeHostAction
                            key="go-to-code-host"
                            repo={this.state.repoOrError}
                            // We need a rev to generate code host URLs, if rev isn't available, we use the default branch or HEAD.
                            rev={
                                this.state.rev ||
                                (!isErrorLike(this.state.repoOrError) &&
                                    this.state.repoOrError.defaultBranch &&
                                    this.state.repoOrError.defaultBranch.displayName) ||
                                'HEAD'
                            }
                            filePath={filePath}
                            commitRange={commitRange}
                            position={position}
                            range={range}
                            externalLinks={this.state.externalLinks}
                        />
                    }
                    {...this.state.repoHeaderContributionsLifecycleProps}
                />
                <ErrorBoundary location={this.props.location}>
                    {this.state.repoOrError.enabled || isSettingsPage ? (
                        <Switch>
                            {[
                                '',
                                `@${this.state.rawRev}`, // must exactly match how the rev was encoded in the URL
                                '/-/blob',
                                '/-/tree',
                                '/-/commits',
                            ].map(routePath => (
                                <Route
                                    path={`${repoMatchURL}${routePath}`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={routePath === ''}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <RepoRevContainer
                                            {...routeComponentProps}
                                            {...transferProps}
                                            routes={this.props.repoRevContainerRoutes}
                                            rev={this.state.rev || ''}
                                            resolvedRevOrError={this.state.resolvedRevOrError}
                                            onResolvedRevOrError={this.onResolvedRevOrError}
                                            // must exactly match how the rev was encoded in the URL
                                            routePrefix={`${repoMatchURL}${
                                                this.state.rawRev ? `@${this.state.rawRev}` : ''
                                            }`}
                                        />
                                    )}
                                />
                            ))}
                            <Route
                                path={`${repoMatchURL}/-/commit/:revspec+`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryGitDataContainer repoName={this.state.repoName}>
                                        <RepositoryCommitPage
                                            {...routeComponentProps}
                                            {...transferProps}
                                            onDidUpdateExternalLinks={this.onDidUpdateExternalLinks}
                                        />
                                    </RepositoryGitDataContainer>
                                )}
                            />
                            <Route
                                path={`${repoMatchURL}/-/branches`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryGitDataContainer repoName={this.state.repoName}>
                                        <RepositoryBranchesArea {...routeComponentProps} {...transferProps} />
                                    </RepositoryGitDataContainer>
                                )}
                            />
                            <Route
                                path={`${repoMatchURL}/-/tags`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryGitDataContainer repoName={this.state.repoName}>
                                        <RepositoryReleasesArea {...routeComponentProps} {...transferProps} />
                                    </RepositoryGitDataContainer>
                                )}
                            />
                            <Route
                                path={`${repoMatchURL}/-/compare/:spec*`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryGitDataContainer repoName={this.state.repoName}>
                                        <RepositoryCompareArea {...routeComponentProps} {...transferProps} />
                                    </RepositoryGitDataContainer>
                                )}
                            />
                            <Route
                                path={`${repoMatchURL}/-/stats`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepositoryGitDataContainer repoName={this.state.repoName}>
                                        <RepositoryStatsArea {...routeComponentProps} {...transferProps} />
                                    </RepositoryGitDataContainer>
                                )}
                            />
                            <Route
                                path={`${repoMatchURL}/-/settings`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepoSettingsArea
                                        {...routeComponentProps}
                                        {...transferProps}
                                        onDidUpdateRepository={this.onDidUpdateRepository}
                                    />
                                )}
                            />
                            <Route key="hardcoded-key" component={RepoPageNotFound} />
                        </Switch>
                    ) : (
                        <RepositoryErrorPage
                            repo={this.state.repoOrError.name}
                            repoID={this.state.repoOrError.id}
                            error="disabled"
                            viewerCanAdminister={viewerCanAdminister}
                            onDidUpdateRepository={this.onDidUpdateRepository}
                        />
                    )}
                </ErrorBoundary>
            </div>
        )
    }

    private onDidUpdateRepository = (update: Partial<GQL.IRepository>) => this.repositoryUpdates.next(update)

    private onDidUpdateExternalLinks = (externalLinks: GQL.IExternalLink[] | undefined): void =>
        this.setState({ externalLinks })

    private onResolvedRevOrError = (v: ResolvedRev | ErrorLike | undefined): void => this.revResolves.next(v)

    private onRepoHeaderContributionsLifecyclePropsChange = (lifecycleProps: RepoHeaderContributionsLifecycleProps) =>
        this.setState({ repoHeaderContributionsLifecycleProps: lifecycleProps })
}

/**
 * Parses the URL path (without the leading slash).
 *
 * TODO(sqs): replace with parseBrowserRepoURL?
 *
 * @param repoRevAndRest a string like /my/repo@myrev/-/blob/my/file.txt
 */
function parseURLPath(repoRevAndRest: string): ParsedRepoRev & { rest?: string } {
    const [repoRev, rest] = repoRevAndRest.split('/-/', 2)
    return { ...parseRepoRev(repoRev), rest }
}
