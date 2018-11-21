import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { merge, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { parseBrowserRepoURL } from '.'
import { makeRepoURI, ParsedRepoRev, parseRepoRev, redirectToExternalHost } from '.'
import { WorkspaceRoot } from '../../../shared/src/api/protocol/plainTypes'
import * as GQL from '../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { HeroPage } from '../components/HeroPage'
import { ExtensionsDocumentsProps } from '../extensions/environment/ExtensionsEnvironment'
import {
    ExtensionsControllerProps,
    ExtensionsProps,
    SettingsCascadeProps,
} from '../extensions/ExtensionsClientCommonContext'
import { searchQueryForRepoRev } from '../search'
import { queryUpdates } from '../search/input/QueryInput'
import { refreshSettings } from '../user/settings/backend'
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
        ExtensionsProps,
        ExtensionsDocumentsProps,
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
    private repositoryAdds = new Subject<void>()
    private revResolves = new Subject<ResolvedRev | ErrorLike | undefined>()
    private subscriptions = new Subscription()

    constructor(props: RepoContainerProps) {
        super(props)

        this.state = {
            ...parseURLPath(props.match.params.repoRevAndRest),
        }
    }

    public componentDidMount(): void {
        // Refresh the list of Sourcegraph extensions to use. This helps when a
        // language server gets automatically enabled after the first repository
        // in that language is added.
        refreshSettings()
            .toPromise()
            .catch(err => console.error(err))

        const parsedRouteChanges = this.routeMatchChanges.pipe(
            map(({ repoRevAndRest }) => parseURLPath(repoRevAndRest))
        )

        // Fetch repository.
        const repositoryChanges = parsedRouteChanges.pipe(
            map(({ repoPath }) => repoPath),
            distinctUntilChanged()
        )
        this.subscriptions.add(
            merge(
                repositoryChanges,
                this.repositoryAdds.pipe(
                    withLatestFrom(repositoryChanges),
                    map(([, repoPath]) => repoPath)
                )
            )
                .pipe(
                    tap(() => this.setState({ repoOrError: undefined })),
                    switchMap(repoPath =>
                        fetchRepository({ repoPath }).pipe(
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
            parsedRouteChanges.subscribe(({ repoPath, rev, rawRev, rest }) => {
                this.setState({ repoPath, rev, rawRev, rest })

                queryUpdates.next(searchQueryForRepoRev(repoPath, rev))
            })
        )

        this.routeMatchChanges.next(this.props.match.params)

        // Merge in repository updates.
        this.subscriptions.add(
            this.repositoryUpdates.subscribe(update =>
                this.setState(({ repoOrError }) => ({ repoOrError: { ...repoOrError, ...update } as GQL.IRepository }))
            )
        )

        // Update the Sourcegraph extensions environment to reflect the current workspace root.
        this.subscriptions.add(
            this.revResolves
                .pipe(
                    map(resolvedRevOrError => {
                        let roots: WorkspaceRoot[] | null = null
                        if (resolvedRevOrError && !isErrorLike(resolvedRevOrError)) {
                            roots = [
                                {
                                    uri: makeRepoURI({
                                        repoPath: this.state.repoPath,
                                        rev: resolvedRevOrError.commitID,
                                    }),
                                },
                            ]
                        }
                        this.props.extensionsOnRootsChange(roots)
                    })
                )
                .subscribe()
        )
        // Clear the Sourcegraph extensions environment's roots when navigating away.
        this.subscriptions.add(() => this.props.extensionsOnRootsChange(null))
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

        const { repoPath, filePath, commitRange, position, range } = parseBrowserRepoURL(
            location.pathname + location.search + location.hash
        )
        const viewerCanAdminister = !!this.props.authenticatedUser && this.props.authenticatedUser.siteAdmin

        if (isErrorLike(this.state.repoOrError)) {
            // Display error page
            switch (this.state.repoOrError.code) {
                case EREPONOTFOUND:
                    return (
                        <RepositoryErrorPage
                            repo={repoPath}
                            repoID={null}
                            error={this.state.repoOrError}
                            viewerCanAdminister={viewerCanAdminister}
                            onDidAddRepository={this.onDidAddRepository}
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
            extensionsOnRootsChange: this.props.extensionsOnRootsChange,
            extensionsOnVisibleTextDocumentsChange: this.props.extensionsOnVisibleTextDocumentsChange,
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
                                <RepositoryGitDataContainer repoPath={this.state.repoPath}>
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
                                <RepositoryGitDataContainer repoPath={this.state.repoPath}>
                                    <RepositoryBranchesArea {...routeComponentProps} {...transferProps} />
                                </RepositoryGitDataContainer>
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/tags`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGitDataContainer repoPath={this.state.repoPath}>
                                    <RepositoryReleasesArea {...routeComponentProps} {...transferProps} />
                                </RepositoryGitDataContainer>
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/compare/:spec*`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGitDataContainer repoPath={this.state.repoPath}>
                                    <RepositoryCompareArea {...routeComponentProps} {...transferProps} />
                                </RepositoryGitDataContainer>
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/stats`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryGitDataContainer repoPath={this.state.repoPath}>
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
            </div>
        )
    }

    private onDidUpdateRepository = (update: Partial<GQL.IRepository>) => this.repositoryUpdates.next(update)
    private onDidAddRepository = () => this.repositoryAdds.next()

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
