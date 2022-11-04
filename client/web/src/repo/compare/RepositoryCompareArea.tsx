import * as React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { HoverMerged } from '@sourcegraph/client-api'
import { HoveredToken, createHoverifier, Hoverifier, HoverState } from '@sourcegraph/codeintellify'
import { isDefined, property } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    FileSpec,
    ModeSpec,
    UIPositionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
} from '@sourcegraph/shared/src/util/url'
import { Alert, LoadingSpinner } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../backend/features'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { WebHoverOverlay } from '../../components/shared'
import { RepositoryFields, Scalars } from '../../graphql-operations'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { RepositoryCompareHeader } from './RepositoryCompareHeader'
import { RepositoryCompareOverviewPage } from './RepositoryCompareOverviewPage'

import styles from './RepositoryCompareArea.module.scss'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
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
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        BreadcrumbSetters {
    repo?: RepositoryFields
    history: H.History
}

interface State extends HoverState<HoverContext, HoverMerged, ActionItemAction> {
    error?: string
}

/**
 * Properties passed to all page components in the repository compare area.
 */
export interface RepositoryCompareAreaPageProps extends PlatformContextProps {
    /** The repository being compared. */
    repo: RepositoryFields

    /** The base of the comparison. */
    base: { repoName: string; repoID: Scalars['ID']; revision?: string | null }

    /** The head of the comparison. */
    head: { repoName: string; repoID: Scalars['ID']; revision?: string | null }

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
    private nextOverlayElement = (element: HTMLElement | null): void => this.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private repositoryCompareAreaElements = new Subject<HTMLElement | null>()
    private nextRepositoryCompareAreaElement = (element: HTMLElement | null): void =>
        this.repositoryCompareAreaElements.next(element)

    private subscriptions = new Subscription()
    private hoverifier: Hoverifier<
        RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
        HoverMerged,
        ActionItemAction
    >

    constructor(props: RepositoryCompareAreaProps) {
        super(props)
        this.hoverifier = createHoverifier<
            RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
            HoverMerged,
            ActionItemAction
        >({
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.repositoryCompareAreaElements),
                map(([, hoverOverlayElement, repositoryCompareAreaElement]) => ({
                    hoverOverlayElement,
                    // The root component element is guaranteed to be rendered after a componentDidUpdate
                    relativeElement: repositoryCompareAreaElement!,
                })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(property('hoverOverlayElement', isDefined))
            ),
            getHover: hoveredToken => getHover(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
            getDocumentHighlights: hoveredToken =>
                getDocumentHighlights(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
            getActions: context => getHoverActions(this.props, context),
        })
        this.subscriptions.add(this.hoverifier)
        this.state = this.hoverifier.hoverState
        this.subscriptions.add(this.hoverifier.hoverStateUpdates.subscribe(update => this.setState(update)))
    }

    private getLSPTextDocumentPositionParams(
        hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
    ): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
        return {
            repoName: hoveredToken.repoName,
            revision: hoveredToken.revision,
            filePath: hoveredToken.filePath,
            commitID: hoveredToken.commitID,
            position: hoveredToken,
            mode: getModeFromPath(hoveredToken.filePath || ''),
        }
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
        this.subscriptions.add(this.props.setBreadcrumb({ key: 'compare', element: <>Compare</> }))
    }

    public shouldComponentUpdate(nextProps: Readonly<RepositoryCompareAreaProps>, nextState: Readonly<State>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { extensionsController } = this.props

        if (this.state.error) {
            return (
                <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={this.state.error} />} />
            )
        }

        let spec: { base: string | null; head: string | null } | null | undefined
        if (this.props.match.params.spec) {
            spec = parseComparisonSpec(decodeURIComponent(this.props.match.params.spec))
        }

        if (!this.props.repo) {
            return <LoadingSpinner />
        }

        const commonProps: RepositoryCompareAreaPageProps = {
            repo: this.props.repo,
            base: { repoID: this.props.repo.id, repoName: this.props.repo.name, revision: spec?.base },
            head: { repoID: this.props.repo.id, repoName: this.props.repo.name, revision: spec?.head },
            routePrefix: this.props.match.url,
            platformContext: this.props.platformContext,
        }
        return (
            <div
                className={classNames('container', styles.repositoryCompareArea)}
                ref={this.nextRepositoryCompareAreaElement}
            >
                <RepositoryCompareHeader className="my-3" {...commonProps} />
                {spec === null ? (
                    <Alert variant="danger">Invalid comparison specifier</Alert>
                ) : (
                    <Switch>
                        <Route
                            path={`${this.props.match.url}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            render={routeComponentProps => (
                                <RepositoryCompareOverviewPage
                                    {...routeComponentProps}
                                    {...commonProps}
                                    hoverifier={this.hoverifier}
                                    isLightTheme={this.props.isLightTheme}
                                    extensionsController={this.props.extensionsController}
                                />
                            )}
                        />
                        <Route key="hardcoded-key" component={NotFoundPage} />
                    </Switch>
                )}
                {this.state.hoverOverlayProps && extensionsController !== null && (
                    <WebHoverOverlay
                        {...this.props}
                        {...this.state.hoverOverlayProps}
                        extensionsController={extensionsController}
                        nav={url => this.props.history.push(url)}
                        hoveredTokenElement={this.state.hoveredTokenElement}
                        telemetryService={this.props.telemetryService}
                        hoverRef={this.nextOverlayElement}
                    />
                )}
            </div>
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
