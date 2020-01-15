import { isEqual, upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { UncontrolledPopover } from 'reactstrap'
import { defer, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, retryWhen, switchMap, tap } from 'rxjs/operators'
import { CloneInProgressError, ECLONEINPROGESS, EREPONOTFOUND, EREVNOTFOUND } from '../../../shared/src/backend/errors'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { HeroPage } from '../components/HeroPage'
import { ChromeExtensionToast } from '../marketing/BrowserExtensionToast'
import { IS_CHROME } from '../marketing/util'
import { ThemeProps } from '../../../shared/src/theme'
import { EventLoggerProps } from '../tracking/eventLogger'
import { RouteDescriptor } from '../util/contributions'
import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ResolvedRev, resolveRev } from './backend'
import { RepoContainerContext } from './RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { EmptyRepositoryPage, RepositoryCloningInProgressPage } from './RepositoryGitDataContainer'
import { RevisionsPopover } from './RevisionsPopover'
import { PatternTypeProps } from '../search'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarItem } from './settings/RepoSettingsSidebar'

/** Props passed to sub-routes of {@link RepoRevContainer}. */
export interface RepoRevContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        EventLoggerProps,
        ActivationProps,
        Pick<
            RepoContainerContext,
            Exclude<keyof RepoContainerContext, 'onDidUpdateRepository' | 'onDidUpdateExternalLinks'>
        >,
        PatternTypeProps {
    repo: GQL.IRepository
    rev: string
    resolvedRev: ResolvedRev

    /** The URL route match for {@link RepoRevContainer}. */
    routePrefix: string
}

/** A sub-route of {@link RepoRevContainer}. */
export interface RepoRevContainerRoute extends RouteDescriptor<RepoRevContainerContext> {}

interface RepoRevContainerProps
    extends RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        EventLoggerProps,
        ExtensionsControllerProps,
        ThemeProps,
        ActivationProps,
        PatternTypeProps {
    routes: readonly RepoRevContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarItems: readonly RepoSettingsSideBarItem[]
    repo: GQL.IRepository
    rev: string
    authenticatedUser: GQL.IUser | null
    routePrefix: string

    /**
     * The resolved rev or an error if it could not be resolved. This value lives in RepoContainer (that
     * component's parent) but originates from this component.
     */
    resolvedRevOrError?: ResolvedRev | ErrorLike

    /** Called when the resolvedRevOrError state in this component's parent should be updated. */
    onResolvedRevOrError: (v: ResolvedRev | ErrorLike | undefined) => void
}

interface RepoRevContainerState {}

/**
 * A container for a repository page that incorporates revisioned Git data. (For example,
 * blob and tree pages are revisioned, but the repository settings page is not.)
 */
export class RepoRevContainer extends React.PureComponent<RepoRevContainerProps, RepoRevContainerState> {
    public state: RepoRevContainerState = {}

    private propsUpdates = new Subject<RepoRevContainerProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const repoRevChanges = that.propsUpdates.pipe(
            // Pick repoName and rev out of the props
            map(props => ({ repoName: props.repo.name, rev: props.rev })),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )

        // Fetch repository revision.
        that.subscriptions.add(
            repoRevChanges
                .pipe(
                    // Reset resolved rev / error state
                    tap(() => that.props.onResolvedRevOrError(undefined)),
                    switchMap(({ repoName, rev }) =>
                        defer(() => resolveRev({ repoName, rev })).pipe(
                            // On a CloneInProgress error, retry after 1s
                            retryWhen(errors =>
                                errors.pipe(
                                    tap(error => {
                                        switch (error.code) {
                                            case ECLONEINPROGESS:
                                                // Display cloning screen to the user and retry
                                                that.props.onResolvedRevOrError(error)
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
                                that.props.onResolvedRevOrError(error)
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    resolvedRev => {
                        that.props.onResolvedRevOrError(resolvedRev)
                    },
                    error => {
                        // Should never be reached because errors are caught above
                        console.error(error)
                    }
                )
        )

        that.propsUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.propsUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!that.props.resolvedRevOrError) {
            // Render nothing while loading
            return null
        }

        if (isErrorLike(that.props.resolvedRevOrError)) {
            // Show error page
            switch (that.props.resolvedRevOrError.code) {
                case ECLONEINPROGESS:
                    return (
                        <RepositoryCloningInProgressPage
                            repoName={that.props.repo.name}
                            progress={(that.props.resolvedRevOrError as CloneInProgressError).progress}
                        />
                    )
                case EREPONOTFOUND:
                    return (
                        <HeroPage
                            icon={MapSearchIcon}
                            title="404: Not Found"
                            subtitle="The requested repository was not found."
                        />
                    )
                case EREVNOTFOUND:
                    if (!that.props.rev) {
                        return <EmptyRepositoryPage />
                    }

                    return (
                        <HeroPage
                            icon={MapSearchIcon}
                            title="404: Not Found"
                            subtitle="The requested revision was not found."
                        />
                    )
                default:
                    return (
                        <HeroPage
                            icon={AlertCircleIcon}
                            title="Error"
                            subtitle={upperFirst(that.props.resolvedRevOrError.message)}
                        />
                    )
            }
        }

        const context: RepoRevContainerContext = {
            platformContext: that.props.platformContext,
            extensionsController: that.props.extensionsController,
            isLightTheme: that.props.isLightTheme,
            telemetryService: that.props.telemetryService,
            activation: that.props.activation,
            repo: that.props.repo,
            repoHeaderContributionsLifecycleProps: that.props.repoHeaderContributionsLifecycleProps,
            resolvedRev: that.props.resolvedRevOrError,
            rev: that.props.rev,
            routePrefix: that.props.routePrefix,
            authenticatedUser: that.props.authenticatedUser,
            settingsCascade: that.props.settingsCascade,
            patternType: that.props.patternType,
            setPatternType: that.props.setPatternType,
            repoSettingsAreaRoutes: that.props.repoSettingsAreaRoutes,
            repoSettingsSidebarItems: that.props.repoSettingsSidebarItems,
        }

        return (
            <div className="repo-rev-container">
                {IS_CHROME && <ChromeExtensionToast />}
                <RepoHeaderContributionPortal
                    position="nav"
                    priority={100}
                    element={
                        <div className="d-flex align-items-center" key="repo-rev">
                            <span className="e2e-revision">
                                {(that.props.rev && that.props.rev === that.props.resolvedRevOrError.commitID
                                    ? that.props.resolvedRevOrError.commitID.slice(0, 7)
                                    : that.props.rev) ||
                                    that.props.resolvedRevOrError.defaultBranch ||
                                    'HEAD'}
                            </span>
                            <button type="button" id="repo-rev-popover" className="btn btn-link px-0">
                                <MenuDownIcon className="icon-inline" />
                            </button>
                            <UncontrolledPopover placement="bottom-start" target="repo-rev-popover" trigger="legacy">
                                <RevisionsPopover
                                    repo={that.props.repo.id}
                                    repoName={that.props.repo.name}
                                    defaultBranch={that.props.resolvedRevOrError.defaultBranch}
                                    currentRev={that.props.rev}
                                    currentCommitID={that.props.resolvedRevOrError.commitID}
                                    history={that.props.history}
                                    location={that.props.location}
                                />
                            </UncontrolledPopover>
                        </div>
                    }
                    repoHeaderContributionsLifecycleProps={that.props.repoHeaderContributionsLifecycleProps}
                />
                <Switch>
                    {/* eslint-disable react/jsx-no-bind */}
                    {that.props.routes.map(
                        ({ path, render, exact, condition = () => true }) =>
                            condition(context) && (
                                <Route
                                    path={that.props.routePrefix + path}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={exact}
                                    render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                />
                            )
                    )}
                    {/* eslint-enable react/jsx-no-bind */}
                </Switch>
                <RepoHeaderContributionPortal
                    position="left"
                    element={<CopyLinkAction key="copy-link" location={that.props.location} />}
                    repoHeaderContributionsLifecycleProps={that.props.repoHeaderContributionsLifecycleProps}
                />
                <RepoHeaderContributionPortal
                    position="right"
                    priority={3}
                    element={
                        <GoToPermalinkAction
                            key="go-to-permalink"
                            rev={that.props.rev}
                            commitID={that.props.resolvedRevOrError.commitID}
                            location={that.props.location}
                            history={that.props.history}
                        />
                    }
                    repoHeaderContributionsLifecycleProps={that.props.repoHeaderContributionsLifecycleProps}
                />
            </div>
        )
    }
}
