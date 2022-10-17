import * as React from 'react'
import { useCallback, useEffect, useMemo, useRef } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ReplaySubject, Subject } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { HoveredToken, createHoverifier } from '@sourcegraph/codeintellify'
import { isDefined, property } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
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
import { Alert, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../backend/features'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { WebHoverOverlay } from '../../components/shared'
import { RepositoryFields, Scalars } from '../../graphql-operations'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
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
export function RepositoryCompareArea(props: RepositoryCompareAreaProps): JSX.Element {
    const compareAreaElement = useRef<HTMLDivElement | null>(null)

    // Element reference subjects passed to `hoverifier`
    const compareAreaElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    useEffect(() => compareAreaElements.next(compareAreaElement.current), [compareAreaElement, compareAreaElements])
    const hoverOverlayElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextOverlayElement = useCallback(
        (overlayElement: HTMLElement | null) => hoverOverlayElements.next(overlayElement),
        [hoverOverlayElements]
    )

    // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
    // reposition hover overlay if needed when `SearchNotebook` rerenders
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => rerenders.next())

    const { extensionsController, platformContext } = props

    const hoverifier = useMemo(
        () =>
            createHoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>({
                hoverOverlayElements,
                hoverOverlayRerenders: rerenders.pipe(
                    withLatestFrom(hoverOverlayElements, compareAreaElements),
                    map(([, hoverOverlayElement, blockElement]) => ({
                        hoverOverlayElement,
                        relativeElement: blockElement,
                    })),
                    filter(property('relativeElement', isDefined)),
                    // Can't reposition HoverOverlay if it wasn't rendered
                    filter(property('hoverOverlayElement', isDefined))
                ),
                getHover: context =>
                    getHover(getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)), {
                        extensionsController,
                    }),
                getDocumentHighlights: context =>
                    getDocumentHighlights(
                        getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)),
                        { extensionsController }
                    ),
                getActions: context => getHoverActions({ extensionsController, platformContext }, context),
            }),
        [hoverOverlayElements, rerenders, compareAreaElements, extensionsController, platformContext]
    )

    // Passed to HoverOverlay
    const hoverState = useObservable(hoverifier.hoverStateUpdates) || {}

    // Dispose hoverifier or change/unmount.
    useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

    let spec: { base: string | null; head: string | null } | null | undefined
    if (props.match.params.spec) {
        spec = parseComparisonSpec(decodeURIComponent(props.match.params.spec))
    }

    // Parse out the optional filePath search param, which is used to show only a single file in the compare view
    const searchParams = new URLSearchParams(props.location.search)
    const path = searchParams.get('filePath')

    props
        .useBreadcrumb(
            useMemo(() => {
                if (props.repo && spec?.head && path) {
                    return {
                        key: 'filePath',
                        className: 'flex-shrink-past-contents',
                        element: (
                            <FilePathBreadcrumbs
                                key="path"
                                repoName={props.repo?.name}
                                revision={spec?.head}
                                filePath={path}
                                isDir={false}
                                telemetryService={props.telemetryService}
                            />
                        ),
                    }
                }

                return null
            }, [props, spec, path])
        )
        .useBreadcrumb(useMemo(() => ({ key: 'compare', element: <>Compare</> }), []))

    if (!props.repo) {
        return <LoadingSpinner />
    }

    const commonProps: RepositoryCompareAreaPageProps = {
        repo: props.repo,
        base: { repoID: props.repo.id, repoName: props.repo.name, revision: spec?.base },
        head: { repoID: props.repo.id, repoName: props.repo.name, revision: spec?.head },
        routePrefix: props.match.url,
        platformContext: props.platformContext,
    }

    return (
        <div className={classNames('container', styles.repositoryCompareArea)} ref={compareAreaElement}>
            <RepositoryCompareHeader className="my-3" {...commonProps} />
            {spec === null ? (
                <Alert variant="danger">Invalid comparison specifier</Alert>
            ) : (
                <Switch>
                    <Route
                        path={`${props.match.url}`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        render={routeComponentProps => (
                            <RepositoryCompareOverviewPage
                                {...routeComponentProps}
                                {...commonProps}
                                path={path}
                                hoverifier={hoverifier}
                                isLightTheme={props.isLightTheme}
                                extensionsController={props.extensionsController}
                            />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            )}
            {hoverState.hoverOverlayProps && extensionsController !== null && (
                <WebHoverOverlay
                    {...props}
                    {...hoverState.hoverOverlayProps}
                    extensionsController={extensionsController}
                    nav={url => props.history.push(url)}
                    hoveredTokenElement={hoverState.hoveredTokenElement}
                    telemetryService={props.telemetryService}
                    hoverRef={nextOverlayElement}
                />
            )}
        </div>
    )
}

export function getLSPTextDocumentPositionParameters(
    position: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
    mode: string
): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: position.repoName,
        filePath: position.filePath,
        commitID: position.commitID,
        revision: position.revision,
        mode,
        position,
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
