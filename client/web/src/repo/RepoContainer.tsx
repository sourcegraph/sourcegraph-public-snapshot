import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { escapeRegExp } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { NEVER, ObservableInput, of } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, encodeURIPathComponent, repeatUntil } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import {
    isCloneInProgressErrorLike,
    isRepoNotFoundErrorLike,
    isRepoSeeOtherErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { escapeSpaces } from '@sourcegraph/shared/src/search/query/filters'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import {
    Icon,
    Button,
    ButtonGroup,
    useLocalStorage,
    useObservable,
    Link,
    Popover,
    PopoverContent,
    Position,
    PopoverTrigger,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { CodeIntelligenceProps } from '../codeintel'
import { BreadcrumbSetters, BreadcrumbsProps } from '../components/Breadcrumbs'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { ActionItemsBarProps, useWebActionItems } from '../extensions/components/ActionItemsBar'
import { ExternalLinkFields, RepositoryFields } from '../graphql-operations'
import { CodeInsightsProps } from '../insights/types'
import { searchQueryForRepoRevision, SearchStreamingProps } from '../search'
import { useNavbarQueryState } from '../stores'
import { useIsBrowserExtensionActiveUser } from '../tracking/BrowserExtensionTracker'
import { RouteDescriptor } from '../util/contributions'
import { parseBrowserRepoURL } from '../util/url'

import { GoToCodeHostAction } from './actions/GoToCodeHostAction'
import type { ExtensionAlertProps } from './actions/InstallIntegrationsAlert'
import { fetchFileExternalLinks, fetchRepository, resolveRevision } from './backend'
import { RepoHeader, RepoHeaderActionButton, RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RepoRevisionContainer, RepoRevisionContainerRoute } from './RepoRevisionContainer'
import { RepositoriesPopover } from './RepositoriesPopover'
import { RepositoryNotFoundPage } from './RepositoryNotFoundPage'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './settings/RepoSettingsSidebar'

import { redirectToExternalHost } from '.'

import styles from './RepoContainer.module.scss'

/**
 * Props passed to sub-routes of {@link RepoContainer}.
 */
export interface RepoContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        HoverThresholdProps,
        TelemetryProps,
        ActivationProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        ExtensionAlertProps {
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]

    /** The URL route match for {@link RepoContainer}. */
    routePrefix: string

    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void

    globbing: boolean

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

/** A sub-route of {@link RepoContainer}. */
export interface RepoContainerRoute extends RouteDescriptor<RepoContainerContext> {}

const RepoPageNotFound: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="The repository page was not found." />
)

interface RepoContainerProps
    extends RouteComponentProps<{ repoRevAndRest: string }>,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ActivationProps,
        ThemeProps,
        ExtensionAlertProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        BreadcrumbSetters,
        BreadcrumbsProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps {
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[]
    repoHeaderActionButtons: readonly RepoHeaderActionButton[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    authenticatedUser: AuthenticatedUser | null
    history: H.History
    globbing: boolean
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
}

export const HOVER_COUNT_KEY = 'hover-count'

export const HOVER_THRESHOLD = 5

export interface HoverThresholdProps {
    /**
     * Called when a hover with content is shown.
     */
    onHoverShown?: () => void
}

/**
 * Renders a horizontal bar and content for a repository page.
 */
export const RepoContainer: React.FunctionComponent<React.PropsWithChildren<RepoContainerProps>> = props => {
    const { repoName, revision, rawRevision, filePath, commitRange, position, range } = parseBrowserRepoURL(
        location.pathname + location.search + location.hash
    )

    // Fetch repository upon mounting the component.
    const repoOrError = useObservable(
        useMemo(
            () =>
                fetchRepository({ repoName }).pipe(
                    catchError(
                        (error): ObservableInput<ErrorLike> => {
                            const redirect = isRepoSeeOtherErrorLike(error)
                            if (redirect) {
                                redirectToExternalHost(redirect)
                                return NEVER
                            }
                            return of(asError(error))
                        }
                    )
                ),
            [repoName]
        )
    )

    const resolvedRevisionOrError = useObservable(
        useMemo(
            () =>
                of(undefined)
                    .pipe(
                        // Wrap in switchMap so we don't break the observable chain when
                        // catchError returns a new observable, so repeatUntil will
                        // properly resubscribe to the outer observable and re-fetch.
                        switchMap(() =>
                            resolveRevision({ repoName, revision }).pipe(
                                catchError(error => {
                                    if (isCloneInProgressErrorLike(error)) {
                                        return of<ErrorLike>(asError(error))
                                    }
                                    throw error
                                })
                            )
                        )
                    )
                    .pipe(
                        repeatUntil(value => !isCloneInProgressErrorLike(value), { delay: 1000 }),
                        catchError(error => of<ErrorLike>(asError(error)))
                    ),
            [repoName, revision]
        )
    )

    // The external links to show in the repository header, if any.
    const [externalLinks, setExternalLinks] = useState<ExternalLinkFields[] | undefined>()

    // The lifecycle props for repo header contributions.
    const [
        repoHeaderContributionsLifecycleProps,
        setRepoHeaderContributionsLifecycleProps,
    ] = useState<RepoHeaderContributionsLifecycleProps>()

    const childBreadcrumbSetters = props.useBreadcrumb(
        useMemo(() => {
            if (isErrorLike(repoOrError) || !repoOrError) {
                return
            }

            return {
                key: 'repository',
                element: (
                    <Popover>
                        <ButtonGroup className="d-inline-flex">
                            <Button
                                to={
                                    resolvedRevisionOrError && !isErrorLike(resolvedRevisionOrError)
                                        ? resolvedRevisionOrError.rootTreeURL
                                        : repoOrError.url
                                }
                                className="text-nowrap test-repo-header-repo-link"
                                variant="secondary"
                                outline={true}
                                size="sm"
                                as={Link}
                            >
                                <Icon as={SourceRepositoryIcon} /> {displayRepoName(repoOrError.name)}
                            </Button>
                            <PopoverTrigger
                                as={Button}
                                className={styles.repoChange}
                                aria-label="Change repository"
                                outline={true}
                                variant="secondary"
                                size="sm"
                            >
                                <Icon as={ChevronDownIcon} />
                            </PopoverTrigger>
                        </ButtonGroup>
                        <PopoverContent position={Position.bottomStart} className="pt-0 pb-0">
                            <RepositoriesPopover
                                currentRepo={repoOrError.id}
                                telemetryService={props.telemetryService}
                            />
                        </PopoverContent>
                    </Popover>
                ),
            }
        }, [repoOrError, resolvedRevisionOrError, props.telemetryService])
    )

    // Update the workspace roots service to reflect the current repo / resolved revision
    useEffect(() => {
        const workspaceRootUri =
            resolvedRevisionOrError &&
            !isErrorLike(resolvedRevisionOrError) &&
            makeRepoURI({
                repoName,
                revision: resolvedRevisionOrError.commitID,
            })

        if (workspaceRootUri) {
            props.extensionsController.extHostAPI
                .then(extensionHostAPI =>
                    extensionHostAPI.addWorkspaceRoot({
                        uri: workspaceRootUri,
                        inputRevision: revision || '',
                    })
                )
                .catch(error => {
                    console.error('Error adding workspace root', error)
                })
        }

        // Clear the Sourcegraph extensions model's roots when navigating away.
        return () => {
            if (workspaceRootUri) {
                props.extensionsController.extHostAPI
                    .then(extensionHostAPI => extensionHostAPI.removeWorkspaceRoot(workspaceRootUri))
                    .catch(error => {
                        console.error('Error removing workspace root', error)
                    })
            }
        }
    }, [props.extensionsController, repoName, resolvedRevisionOrError, revision])

    // Update the navbar query to reflect the current repo / revision
    const { globbing } = props
    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)
    useEffect(() => {
        let query = searchQueryForRepoRevision(repoName, globbing, revision)
        if (filePath) {
            query = `${query.trimEnd()} file:${escapeSpaces(globbing ? filePath : '^' + escapeRegExp(filePath))}`
        }
        onNavbarQueryChange({
            query,
        })
    }, [revision, filePath, repoName, onNavbarQueryChange, globbing])

    const { useActionItemsBar, useActionItemsToggle } = useWebActionItems()

    const codeHostIntegrationMessaging =
        (!isErrorLike(props.settingsCascade.final) &&
            props.settingsCascade.final?.['alerts.codeHostIntegrationMessaging']) ||
        'browser-extension'
    const isBrowserExtensionActiveUser = useIsBrowserExtensionActiveUser()
    // Browser extension discoverability features (alert, popover for `GoToCodeHostAction)
    const [hasDismissedPopover, setHasDismissedPopover] = useState(false)
    const [hoverCount, setHoverCount] = useLocalStorage(HOVER_COUNT_KEY, 0)
    const canShowPopover =
        !hasDismissedPopover &&
        isBrowserExtensionActiveUser === false &&
        codeHostIntegrationMessaging === 'browser-extension' &&
        hoverCount >= HOVER_THRESHOLD

    // Increment hovers that the user has seen. Enable browser extension discoverability
    // features after hover count threshold is reached (e.g. alerts, popovers)
    // Store hover count in ref to avoid circular dependency
    // hoverCount -> onHoverShown -> WebHoverOverlay (onHoverShown in useEffect deps) -> onHoverShown()
    const hoverCountReference = useRef<number>(hoverCount)
    hoverCountReference.current = hoverCount
    const onHoverShown = useCallback(() => {
        const count = hoverCountReference.current + 1
        if (count > HOVER_THRESHOLD) {
            // No need to keep updating localStorage
            return
        }
        setHoverCount(count)
    }, [setHoverCount])

    const onPopoverDismissed = useCallback(() => {
        setHasDismissedPopover(true)
    }, [])

    if (!repoOrError) {
        // Render nothing while loading
        return null
    }

    const viewerCanAdminister = !!props.authenticatedUser && props.authenticatedUser.siteAdmin

    if (isErrorLike(repoOrError)) {
        // Display error page
        if (isRepoNotFoundErrorLike(repoOrError)) {
            return <RepositoryNotFoundPage repo={repoName} viewerCanAdminister={viewerCanAdminister} />
        }
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={repoOrError} />} />
    }

    const isCodeIntelRepositoryBadgeEnabled =
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final?.experimentalFeatures?.codeIntelRepositoryBadge?.enabled === true

    // Remove leading repository name and possible leading revision, then compare the remaining routes to
    // see if we should display the code intelligence badge for this route. We want this to be visible on
    // the repo root page, as well as directory and code views, but not administrative/non-code views.
    const matchRevisionAndRest = props.match.params.repoRevAndRest.slice(repoName.length)
    const matchOnlyRest =
        revision && matchRevisionAndRest.startsWith(`@${revision || ''}`)
            ? matchRevisionAndRest.slice(revision.length + 1)
            : matchRevisionAndRest
    const isCodeIntelRepositoryBadgeVisibleOnRoute =
        matchOnlyRest === '' || matchOnlyRest.startsWith('/-/tree') || matchOnlyRest.startsWith('/-/blob')

    const repoMatchURL = '/' + encodeURIPathComponent(repoName)

    const context: RepoContainerContext = {
        ...props,
        ...repoHeaderContributionsLifecycleProps,
        ...childBreadcrumbSetters,
        onHoverShown,
        repo: repoOrError,
        routePrefix: repoMatchURL,
        onDidUpdateExternalLinks: setExternalLinks,
        useActionItemsBar,
    }
    return (
        <div className={classNames('w-100 d-flex flex-column', styles.repoContainer)}>
            <RepoHeader
                actionButtons={props.repoHeaderActionButtons}
                useActionItemsToggle={useActionItemsToggle}
                breadcrumbs={props.breadcrumbs}
                revision={revision}
                repo={repoOrError}
                resolvedRev={resolvedRevisionOrError}
                onLifecyclePropsChange={setRepoHeaderContributionsLifecycleProps}
                location={props.location}
                history={props.history}
                settingsCascade={props.settingsCascade}
                authenticatedUser={props.authenticatedUser}
                platformContext={props.platformContext}
                extensionsController={props.extensionsController}
                telemetryService={props.telemetryService}
            />
            <RepoHeaderContributionPortal
                position="right"
                priority={2}
                id="go-to-code-host"
                {...repoHeaderContributionsLifecycleProps}
            >
                {({ actionType }) => (
                    <GoToCodeHostAction
                        key="go-to-code-host"
                        repo={repoOrError}
                        // We need a revision to generate code host URLs, if revision isn't available, we use the default branch or HEAD.
                        revision={rawRevision || repoOrError.defaultBranch?.displayName || 'HEAD'}
                        filePath={filePath}
                        commitRange={commitRange}
                        position={position}
                        range={range}
                        externalLinks={externalLinks}
                        fetchFileExternalLinks={fetchFileExternalLinks}
                        canShowPopover={canShowPopover}
                        onPopoverDismissed={onPopoverDismissed}
                        actionType={actionType}
                        repoName={repoName}
                    />
                )}
            </RepoHeaderContributionPortal>

            {isCodeIntelRepositoryBadgeEnabled && isCodeIntelRepositoryBadgeVisibleOnRoute && (
                <RepoHeaderContributionPortal
                    position="right"
                    priority={110}
                    id="code-intelligence-status"
                    {...repoHeaderContributionsLifecycleProps}
                >
                    {({ actionType }) =>
                        props.codeIntelligenceBadgeMenu && actionType === 'nav' ? (
                            <props.codeIntelligenceBadgeMenu
                                key="code-intelligence-status"
                                repoName={repoName}
                                revision={rawRevision || 'HEAD'}
                                filePath={filePath || ''}
                                settingsCascade={props.settingsCascade}
                            />
                        ) : (
                            <></>
                        )
                    }
                </RepoHeaderContributionPortal>
            )}

            <ErrorBoundary location={props.location}>
                <Switch>
                    {[
                        '',
                        ...(rawRevision ? [`@${rawRevision}`] : []), // must exactly match how the revision was encoded in the URL
                        '/-/blob',
                        '/-/tree',
                        '/-/commits',
                        '/-/docs',
                        '/-/branch',
                        '/-/contributors',
                        '/-/compare',
                        '/-/tag',
                        '/-/home',
                    ].map(routePath => (
                        <Route
                            path={`${repoMatchURL}${routePath}`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={routePath === ''}
                            render={routeComponentProps => (
                                <RepoRevisionContainer
                                    {...routeComponentProps}
                                    {...context}
                                    {...childBreadcrumbSetters}
                                    routes={props.repoRevisionContainerRoutes}
                                    revision={revision || ''}
                                    resolvedRevisionOrError={resolvedRevisionOrError}
                                    // must exactly match how the revision was encoded in the URL
                                    routePrefix={`${repoMatchURL}${rawRevision ? `@${rawRevision}` : ''}`}
                                    useActionItemsBar={useActionItemsBar}
                                />
                            )}
                        />
                    ))}
                    {props.repoContainerRoutes.map(
                        ({ path, render, exact, condition = () => true }) =>
                            condition(context) && (
                                <Route
                                    path={context.routePrefix + path}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={exact}
                                    // RouteProps.render is an exception
                                    render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                                />
                            )
                    )}
                    <Route key="hardcoded-key" component={RepoPageNotFound} />
                </Switch>
            </ErrorBoundary>
        </div>
    )
}
