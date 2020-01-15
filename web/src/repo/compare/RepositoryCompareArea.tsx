import { createHoverifier, HoveredToken, Hoverifier, HoverState } from '@sourcegraph/codeintellify'
import { isEqual, upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { getHoverActions } from '../../../../shared/src/hover/actions'
import { HoverContext } from '../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { propertyIsDefined } from '../../../../shared/src/util/types'
import {
    escapeRevspecForURL,
    FileSpec,
    ModeSpec,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
} from '../../../../shared/src/util/url'
import { getHover } from '../../backend/features'
import { HeroPage } from '../../components/HeroPage'
import { WebHoverOverlay } from '../../components/shared'
import { EventLoggerProps } from '../../tracking/eventLogger'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { RepositoryCompareHeader } from './RepositoryCompareHeader'
import { RepositoryCompareOverviewPage } from './RepositoryCompareOverviewPage'
import { ThemeProps } from '../../../../shared/src/theme'

const NotFoundPage: React.FunctionComponent = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository comparison page was not found."
    />
)

interface RepositoryCompareAreaProps
    extends RouteComponentProps<{ spec: string }>,
        RepoHeaderContributionsLifecycleProps,
        PlatformContextProps,
        EventLoggerProps,
        ExtensionsControllerProps,
        ThemeProps {
    repo: GQL.IRepository
}

interface State extends HoverState<HoverContext, HoverMerged, ActionItemAction> {
    error?: string
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps extends PlatformContextProps {
    /** The repository being compared. */
    repo: GQL.IRepository

    /** The base of the comparison. */
    base: { repoName: string; repoID: GQL.ID; rev?: string | null }

    /** The head of the comparison. */
    head: { repoName: string; repoID: GQL.ID; rev?: string | null }

    /** The URL route prefix for the comparison. */
    routePrefix: string
}

/**
 * Renders pages related to a repository comparison.
 */
export class RepositoryCompareArea extends React.Component<RepositoryCompareAreaProps, State> {
    private componentUpdates = new Subject<RepositoryCompareAreaProps>()

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null): void => that.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private repositoryCompareAreaElements = new Subject<HTMLElement | null>()
    private nextRepositoryCompareAreaElement = (element: HTMLElement | null): void =>
        that.repositoryCompareAreaElements.next(element)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent): void => that.closeButtonClicks.next(event)

    private subscriptions = new Subscription()
    private hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>

    constructor(props: RepositoryCompareAreaProps) {
        super(props)
        that.hoverifier = createHoverifier<
            RepoSpec & RevSpec & FileSpec & ResolvedRevSpec,
            HoverMerged,
            ActionItemAction
        >({
            closeButtonClicks: that.closeButtonClicks,
            hoverOverlayElements: that.hoverOverlayElements,
            hoverOverlayRerenders: that.componentUpdates.pipe(
                withLatestFrom(that.hoverOverlayElements, that.repositoryCompareAreaElements),
                map(([, hoverOverlayElement, repositoryCompareAreaElement]) => ({
                    hoverOverlayElement,
                    // The root component element is guaranteed to be rendered after a componentDidUpdate
                    relativeElement: repositoryCompareAreaElement!,
                })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            getHover: hoveredToken => getHover(that.getLSPTextDocumentPositionParams(hoveredToken), that.props),
            getActions: context => getHoverActions(that.props, context),
            pinningEnabled: true,
        })
        that.subscriptions.add(that.hoverifier)
        that.state = that.hoverifier.hoverState
        that.subscriptions.add(that.hoverifier.hoverStateUpdates.subscribe(update => that.setState(update)))
    }

    private getLSPTextDocumentPositionParams(
        hoveredToken: HoveredToken & RepoSpec & RevSpec & FileSpec & ResolvedRevSpec
    ): RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec & ModeSpec {
        return {
            repoName: hoveredToken.repoName,
            rev: hoveredToken.rev,
            filePath: hoveredToken.filePath,
            commitID: hoveredToken.commitID,
            position: hoveredToken,
            mode: getModeFromPath(hoveredToken.filePath || ''),
        }
    }

    public componentDidMount(): void {
        that.componentUpdates.next(that.props)
    }

    public shouldComponentUpdate(nextProps: Readonly<RepositoryCompareAreaProps>, nextState: Readonly<State>): boolean {
        return !isEqual(that.props, nextProps) || !isEqual(that.state, nextState)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (that.state.error) {
            return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(that.state.error)} />
        }

        let spec: { base: string | null; head: string | null } | null | undefined
        if (that.props.match.params.spec) {
            spec = parseComparisonSpec(decodeURIComponent(that.props.match.params.spec))
        }

        const commonProps: RepositoryCompareAreaPageProps = {
            repo: that.props.repo,
            base: { repoID: that.props.repo.id, repoName: that.props.repo.name, rev: spec?.base },
            head: { repoID: that.props.repo.id, repoName: that.props.repo.name, rev: spec?.head },
            routePrefix: that.props.match.url,
            platformContext: that.props.platformContext,
        }

        return (
            <div className="repository-compare-area container" ref={that.nextRepositoryCompareAreaElement}>
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="compare">Compare</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={that.props.repoHeaderContributionsLifecycleProps}
                />
                <RepositoryCompareHeader
                    className="my-3"
                    {...commonProps}
                    onUpdateComparisonSpec={that.onUpdateComparisonSpec}
                />
                {spec === null ? (
                    <div className="alert alert-danger">Invalid comparison specifier</div>
                ) : (
                    <Switch>
                        {/* eslint-disable react/jsx-no-bind */}
                        <Route
                            path={`${that.props.match.url}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            render={routeComponentProps => (
                                <RepositoryCompareOverviewPage
                                    {...routeComponentProps}
                                    {...commonProps}
                                    hoverifier={that.hoverifier}
                                    isLightTheme={that.props.isLightTheme}
                                    extensionsController={that.props.extensionsController}
                                />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                        {/* eslint-enable react/jsx-no-bind */}
                    </Switch>
                )}
                {that.state.hoverOverlayProps && (
                    <WebHoverOverlay
                        {...that.props}
                        {...that.state.hoverOverlayProps}
                        telemetryService={that.props.telemetryService}
                        hoverRef={that.nextOverlayElement}
                        onCloseButtonClick={that.nextCloseButtonClick}
                    />
                )}
            </div>
        )
    }

    private onUpdateComparisonSpec = (newBaseSpec: string, newHeadSpec: string): void => {
        that.props.history.push(
            `/${that.props.repo.name}/-/compare${
                newBaseSpec || newHeadSpec
                    ? `/${escapeRevspecForURL(newBaseSpec || '')}...${escapeRevspecForURL(newHeadSpec || '')}`
                    : ''
            }`
        )
    }
}

function parseComparisonSpec(spec: string): { base: string | null; head: string | null } | null {
    if (!spec.includes('...')) {
        return null
    }
    const parts = spec.split('...', 2).map(decodeURIComponent)
    return {
        base: parts[0] || null,
        head: parts[1] || null,
    }
}
