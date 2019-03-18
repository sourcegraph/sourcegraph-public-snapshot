import {
    createHoverifier,
    HoveredToken,
    HoveredTokenContext,
    Hoverifier,
    HoverOverlay,
    HoverState,
} from '@sourcegraph/codeintellify'
import { isEqual, upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link, LinkProps } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'
import { getHover, getJumpURL } from '../../backend/features'
import * as GQL from '../../backend/graphqlschema'
import { LSPTextDocumentPositionParams } from '../../backend/lsp'
import { HeroPage } from '../../components/HeroPage'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps, ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { eventLogger } from '../../tracking/eventLogger'
import { getModeFromPath } from '../../util'
import { propertyIsDefined } from '../../util/types'
import { escapeRevspecForURL } from '../../util/url'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderBreadcrumbNavItem } from '../RepoHeaderBreadcrumbNavItem'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'
import { RepositoryCompareHeader } from './RepositoryCompareHeader'
import { RepositoryCompareOverviewPage } from './RepositoryCompareOverviewPage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested repository comparison page was not found."
    />
)

interface Props
    extends RouteComponentProps<{ spec: string }>,
        RepoHeaderContributionsLifecycleProps,
        ExtensionsProps,
        ExtensionsControllerProps,
        ExtensionsDocumentsProps {
    repo: GQL.IRepository
}

interface State extends HoverState {
    error?: string
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps extends ExtensionsProps {
    /** The repository being compared. */
    repo: GQL.IRepository

    /** The base of the comparison. */
    base: { repoPath: string; repoID: GQL.ID; rev?: string | null }

    /** The head of the comparison. */
    head: { repoPath: string; repoID: GQL.ID; rev?: string | null }

    /** The URL route prefix for the comparison. */
    routePrefix: string
}

const logTelemetryEvent = (event: string, data?: any) => eventLogger.log(event, data)
const LinkComponent = (props: LinkProps) => <Link {...props} />

/**
 * Renders pages related to a repository comparison.
 */
export class RepositoryCompareArea extends React.Component<Props, State> {
    private componentUpdates = new Subject<Props>()

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private repositoryCompareAreaElements = new Subject<HTMLElement | null>()
    private nextRepositoryCompareAreaElement = (element: HTMLElement | null) =>
        this.repositoryCompareAreaElements.next(element)

    /** Emits when the go to definition button was clicked */
    private goToDefinitionClicks = new Subject<MouseEvent>()
    private nextGoToDefinitionClick = (event: MouseEvent) => this.goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent) => this.closeButtonClicks.next(event)

    private subscriptions = new Subscription()
    private hoverifier: Hoverifier

    constructor(props: Props) {
        super(props)
        this.hoverifier = createHoverifier({
            closeButtonClicks: this.closeButtonClicks,
            goToDefinitionClicks: this.goToDefinitionClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.repositoryCompareAreaElements),
                map(([, hoverOverlayElement, repositoryCompareAreaElement]) => ({
                    hoverOverlayElement,
                    // The root component element is guaranteed to be rendered after a componentDidUpdate
                    relativeElement: repositoryCompareAreaElement!,
                })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            pushHistory: path => this.props.history.push(path),
            logTelemetryEvent,
            fetchHover: hoveredToken => getHover(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
            fetchJumpURL: hoveredToken => getJumpURL(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
        })
        this.subscriptions.add(this.hoverifier)
        this.state = this.hoverifier.hoverState
        this.subscriptions.add(this.hoverifier.hoverStateUpdates.subscribe(update => this.setState(update)))
    }

    private getLSPTextDocumentPositionParams(
        hoveredToken: HoveredToken & HoveredTokenContext
    ): LSPTextDocumentPositionParams {
        return {
            repoPath: hoveredToken.repoPath,
            rev: hoveredToken.rev,
            filePath: hoveredToken.filePath,
            commitID: hoveredToken.commitID,
            position: hoveredToken,
            mode: getModeFromPath(hoveredToken.filePath || ''),
        }
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate(nextProps: Readonly<Props>, nextState: Readonly<State>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={upperFirst(this.state.error)} />
        }

        let spec: { base: string | null; head: string | null } | null | undefined
        if (this.props.match.params.spec) {
            spec = parseComparisonSpec(this.props.match.params.spec)
        }

        const commonProps: RepositoryCompareAreaPageProps = {
            repo: this.props.repo,
            base: { repoID: this.props.repo.id, repoPath: this.props.repo.name, rev: spec && spec.base },
            head: { repoID: this.props.repo.id, repoPath: this.props.repo.name, rev: spec && spec.head },
            routePrefix: this.props.match.url,
            extensions: this.props.extensions,
        }

        return (
            <div className="repository-compare-area area--vertical" ref={this.nextRepositoryCompareAreaElement}>
                <RepoHeaderContributionPortal
                    position="nav"
                    element={<RepoHeaderBreadcrumbNavItem key="compare">Compare</RepoHeaderBreadcrumbNavItem>}
                    repoHeaderContributionsLifecycleProps={this.props.repoHeaderContributionsLifecycleProps}
                />
                <RepositoryCompareHeader
                    className="area--vertical__header"
                    {...commonProps}
                    onUpdateComparisonSpec={this.onUpdateComparisonSpec}
                />
                <div className="area--vertical__content">
                    <div className="area--vertical__content-inner">
                        {spec === null ? (
                            <div className="alert alert-danger">Invalid comparison specifier</div>
                        ) : (
                            <Switch>
                                <Route
                                    path={`${this.props.match.url}`}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={true}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    render={routeComponentProps => (
                                        <RepositoryCompareOverviewPage
                                            {...routeComponentProps}
                                            {...commonProps}
                                            hoverifier={this.hoverifier}
                                            extensionsController={this.props.extensionsController}
                                            extensionsOnVisibleTextDocumentsChange={
                                                this.props.extensionsOnVisibleTextDocumentsChange
                                            }
                                        />
                                    )}
                                />
                                <Route key="hardcoded-key" component={NotFoundPage} />
                            </Switch>
                        )}
                    </div>
                </div>
                {this.state.hoverOverlayProps && (
                    <HoverOverlay
                        {...this.state.hoverOverlayProps}
                        logTelemetryEvent={logTelemetryEvent}
                        linkComponent={LinkComponent}
                        hoverRef={this.nextOverlayElement}
                        onGoToDefinitionClick={this.nextGoToDefinitionClick}
                        onCloseButtonClick={this.nextCloseButtonClick}
                    />
                )}
            </div>
        )
    }

    private onUpdateComparisonSpec = (newBaseSpec: string, newHeadSpec: string): void => {
        this.props.history.push(
            `/${this.props.repo.name}/-/compare${
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
