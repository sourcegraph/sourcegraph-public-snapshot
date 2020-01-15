import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription, concat } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap, tap } from 'rxjs/operators'
import { redirectToExternalHost } from '.'
import { EREPONOTFOUND, EREPOSEEOTHER, RepoSeeOtherError } from '../../../shared/src/backend/errors'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike, asError } from '../../../shared/src/util/errors'
import { makeRepoURI } from '../../../shared/src/util/url'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { searchQueryForRepoRev, PatternTypeProps } from '../search'
import { queryUpdates } from '../search/input/QueryInput'
import { EventLoggerProps } from '../tracking/eventLogger'
import { RouteDescriptor } from '../util/contributions'
import { parseBrowserRepoURL, ParsedRepoRev, parseRepoRev } from '../util/url'
import { GoToCodeHostAction } from './actions/GoToCodeHostAction'
import { fetchRepository, ResolvedRev } from './backend'
import { RepoHeader, RepoHeaderActionButton, RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RepoRevContainer, RepoRevContainerRoute } from './RepoRevContainer'
import { RepositoryNotFoundPage } from './RepositoryNotFoundPage'
import { ThemeProps } from '../../../shared/src/theme'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarItem } from './settings/RepoSettingsSidebar'

/**
 * Props passed to sub-routes of {@link RepoContainer}.
 */
export interface RepoContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        EventLoggerProps,
        ActivationProps,
        PatternTypeProps {
    repo: GQL.IRepository
    authenticatedUser: GQL.IUser | null
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarItems: readonly RepoSettingsSideBarItem[]

    /** The URL route match for {@link RepoContainer}. */
    routePrefix: string

    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
    onDidUpdateExternalLinks: (externalLinks: GQL.IExternalLink[] | undefined) => void
}

/** A sub-route of {@link RepoContainer}. */
export interface RepoContainerRoute extends RouteDescriptor<RepoContainerContext> {}

const RepoPageNotFound: React.FunctionComponent = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="The repository page was not found." />
)

interface RepoContainerProps
    extends RouteComponentProps<{ repoRevAndRest: string }>,
        SettingsCascadeProps,
        PlatformContextProps,
        EventLoggerProps,
        ExtensionsControllerProps,
        ActivationProps,
        ThemeProps,
        PatternTypeProps {
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevContainerRoutes: readonly RepoRevContainerRoute[]
    repoHeaderActionButtons: readonly RepoHeaderActionButton[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarItems: readonly RepoSettingsSideBarItem[]
    authenticatedUser: GQL.IUser | null
}

interface RepoRevContainerState extends ParsedRepoRev {
    filePath?: string

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
    private componentUpdates = new Subject<RepoContainerProps>()
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
        const parsedRouteChanges = this.componentUpdates.pipe(
            map(props => props.match.params.repoRevAndRest),
            distinctUntilChanged(),
            map(parseURLPath)
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
                        concat(
                            [undefined],
                            fetchRepository({ repoName }).pipe(
                                catchError(error => {
                                    switch (error.code) {
                                        case EREPOSEEOTHER:
                                            redirectToExternalHost((error as RepoSeeOtherError).redirectURL)
                                            return []
                                    }
                                    return [asError(error)]
                                })
                            )
                        )
                    )
                )
                .subscribe(repoOrError => {
                    this.setState({ repoOrError })
                })
        )

        // Update resolved revision in state
        this.subscriptions.add(this.revResolves.subscribe(resolvedRevOrError => this.setState({ resolvedRevOrError })))

        // Update header and other global state.
        this.subscriptions.add(
            parsedRouteChanges.subscribe(({ repoName, rev, rawRev }) => {
                this.setState({ repoName, rev, rawRev })

                queryUpdates.next(searchQueryForRepoRev(repoName, rev))
            })
        )

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
                        this.props.extensionsController.services.workspace.roots.next(
                            resolvedRevOrError && !isErrorLike(resolvedRevOrError)
                                ? [
                                      {
                                          uri: makeRepoURI({
                                              repoName: this.state.repoName,
                                              rev: resolvedRevOrError.commitID,
                                          }),
                                          inputRevision: this.state.rev || '',
                                      },
                                  ]
                                : []
                        )
                    })
                )
                .subscribe()
        )
        // Clear the Sourcegraph extensions model's roots when navigating away.
        this.subscriptions.add(() => this.props.extensionsController.services.workspace.roots.next([]))

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
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
                    return <RepositoryNotFoundPage repo={repoName} viewerCanAdminister={viewerCanAdminister} />
                default:
                    return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={this.state.repoOrError.message} />
            }
        }

        const repoMatchURL = `/${this.state.repoOrError.name}`

        const context: RepoContainerContext = {
            repo: this.state.repoOrError,
            authenticatedUser: this.props.authenticatedUser,
            isLightTheme: this.props.isLightTheme,
            activation: this.props.activation,
            telemetryService: this.props.telemetryService,
            routePrefix: repoMatchURL,
            settingsCascade: this.props.settingsCascade,
            platformContext: this.props.platformContext,
            extensionsController: this.props.extensionsController,
            ...this.state.repoHeaderContributionsLifecycleProps,
            onDidUpdateExternalLinks: this.onDidUpdateExternalLinks,
            onDidUpdateRepository: this.onDidUpdateRepository,
            patternType: this.props.patternType,
            setPatternType: this.props.setPatternType,
            repoSettingsAreaRoutes: this.props.repoSettingsAreaRoutes,
            repoSettingsSidebarItems: this.props.repoSettingsSidebarItems,
        }

        return (
            <div className="repo-container e2e-repo-container w-100 d-flex flex-column">
                <RepoHeader
                    {...this.props}
                    actionButtons={this.props.repoHeaderActionButtons}
                    rev={this.state.rev}
                    repo={this.state.repoOrError}
                    resolvedRev={this.state.resolvedRevOrError}
                    onLifecyclePropsChange={this.onRepoHeaderContributionsLifecyclePropsChange}
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
                    <Switch>
                        {/* eslint-disable react/jsx-no-bind */}
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
                                render={routeComponentProps => (
                                    <RepoRevContainer
                                        {...routeComponentProps}
                                        {...context}
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
                        {this.props.repoContainerRoutes.map(
                            ({ path, render, exact, condition = () => true }) =>
                                condition(context) && (
                                    <Route
                                        path={context.routePrefix + path}
                                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                        exact={exact}
                                        // RouteProps.render is an exception
                                        // eslint-disable-next-line react/jsx-no-bind
                                        render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                    />
                                )
                        )}
                        <Route key="hardcoded-key" component={RepoPageNotFound} />
                        {/* eslint-enable react/jsx-no-bind */}
                    </Switch>
                </ErrorBoundary>
            </div>
        )
    }

    private onDidUpdateRepository = (update: Partial<GQL.IRepository>): void => this.repositoryUpdates.next(update)

    private onDidUpdateExternalLinks = (externalLinks: GQL.IExternalLink[] | undefined): void =>
        this.setState({ externalLinks })

    private onResolvedRevOrError = (v: ResolvedRev | ErrorLike | undefined): void => this.revResolves.next(v)

    private onRepoHeaderContributionsLifecyclePropsChange = (
        lifecycleProps: RepoHeaderContributionsLifecycleProps
    ): void => this.setState({ repoHeaderContributionsLifecycleProps: lifecycleProps })
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
