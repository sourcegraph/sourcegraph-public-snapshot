import React, { FC, Suspense, useEffect, useMemo, useState } from 'react'

import { mdiSourceRepository } from '@mdi/js'
import classNames from 'classnames'
import { escapeRegExp } from 'lodash'
import { Location, useLocation, Route, Routes } from 'react-router-dom'
import { NEVER, of } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { StreamingSearchResultsListProps } from '@sourcegraph/branded'
import { asError, ErrorLike, isErrorLike, logger, repeatUntil } from '@sourcegraph/common'
import {
    isCloneInProgressErrorLike,
    isRepoSeeOtherErrorLike,
    isRevisionNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { RepoQuestionIcon } from '@sourcegraph/shared/src/components/icons'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { EditorHint, SearchContextProps } from '@sourcegraph/shared/src/search'
import { escapeSpaces } from '@sourcegraph/shared/src/search/query/filters'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Link, Panel, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { CodeIntelligenceProps } from '../codeintel'
import { CodyChat } from '../cody/CodyChat'
import { BreadcrumbSetters, BreadcrumbsProps } from '../components/Breadcrumbs'
import { RouteError } from '../components/ErrorBoundary'
import { HeroPage } from '../components/HeroPage'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { ExternalLinkFields, RepositoryFields } from '../graphql-operations'
import { CodeInsightsProps } from '../insights/types'
import { NotebookProps } from '../notebooks'
import { OwnConfigProps } from '../own/OwnConfigProps'
import { searchQueryForRepoRevision, SearchStreamingProps } from '../search'
import { useExperimentalQueryInput } from '../search/useExperimentalSearchInput'
import { useNavbarQueryState } from '../stores'
import { useChatStore } from '../stores/codyChat'
import { useCodySidebarStore } from '../stores/codySidebar'
import { RouteV6Descriptor } from '../util/contributions'
import { parseBrowserRepoURL } from '../util/url'

import { GoToCodeHostAction } from './actions/GoToCodeHostAction'
import { fetchFileExternalLinks, ResolvedRevision, resolveRepoRevision } from './backend'
import { AskCodyButton } from './cody/AskCodyButton'
import { RepoContainerError } from './RepoContainerError'
import { RepoHeader, RepoHeaderActionButton, RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import {
    RepoRevisionContainer,
    RepoRevisionContainerContext,
    RepoRevisionContainerRoute,
} from './RepoRevisionContainer'
import { repoSplat } from './repoRevisionContainerRoutes'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './settings/RepoSettingsSidebar'
import { repoSettingsAreaPath } from './settings/routes'

import styles from './RepoContainer.module.scss'

const RepoSettingsArea = lazyComponent(() => import('./settings/RepoSettingsArea'), 'RepoSettingsArea')

const CODY_SIDEBAR_SIZES = { default: 350, max: 1200, min: 250 }

/**
 * Props passed to sub-routes of {@link RepoContainer}.
 */
export interface RepoContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        HoverThresholdProps,
        TelemetryProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        BreadcrumbSetters,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps {
    repo: RepositoryFields
    repoName: string
    resolvedRevisionOrError: ResolvedRevision | ErrorLike | undefined
    authenticatedUser: AuthenticatedUser | null
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]

    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

/**
 * Props passed to sub-routes of {@link RepoContainer} which are specific to repository settings.
 */
export interface RepoSettingsContainerContext extends Omit<RepoContainerContext, 'repo' | 'resolvedRevisionOrError'> {}

/** A sub-route of {@link RepoContainer}. */
export interface RepoContainerRoute extends RouteV6Descriptor<RepoContainerContext> {}

interface RepoContainerProps
    extends SettingsCascadeProps<Settings>,
        PlatformContextProps,
        TelemetryProps,
        ExtensionsControllerProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        BreadcrumbSetters,
        BreadcrumbsProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        NotebookProps,
        OwnConfigProps {
    repoContainerRoutes: readonly RepoContainerRoute[]
    repoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[]
    repoHeaderActionButtons: readonly RepoHeaderActionButton[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    authenticatedUser: AuthenticatedUser | null
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
}

export interface HoverThresholdProps {
    /**
     * Called when a hover with content is shown.
     */
    onHoverShown?: () => void
}

/**
 * Renders a horizontal bar and content for a repository page.
 */
export const RepoContainer: FC<RepoContainerProps> = props => {
    const { extensionsController, repoContainerRoutes, authenticatedUser, selectedSearchContextSpec } = props

    const location = useLocation()

    const { repoName, revision, rawRevision, filePath, commitRange, position, range } = parseBrowserRepoURL(
        location.pathname + location.search + location.hash
    )

    const resolvedRevisionOrError = useObservable(
        useMemo(
            () =>
                of(undefined)
                    .pipe(
                        // Wrap in switchMap, so we don't break the observable chain when
                        // catchError returns a new observable, so repeatUntil will
                        // properly resubscribe to the outer observable and re-fetch.
                        switchMap(() =>
                            resolveRepoRevision({ repoName, revision }).pipe(
                                catchError(error => {
                                    const redirect = isRepoSeeOtherErrorLike(error)

                                    if (redirect) {
                                        redirectToExternalHost(redirect)
                                        return NEVER
                                    }

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

    const [isCodyEnabled] = useFeatureFlag('cody-experimental')
    const focusCodyShortcut = useKeyboardShortcut('focusCody')
    const {
        isOpen: isCodySidebarOpen,
        setIsOpen: setIsCodySidebarOpen,
        onResize: onCodySidebarResize,
    } = useCodySidebarStore()
    // TODO: This hook call is used to initialize the chat store with the right repo name.
    useChatStore({ codebase: repoName, setIsCodySidebarOpen })

    /**
     * A long time ago, we fetched `repo` in a separate GraphQL query.
     * This GraphQL query was merged into the `resolveRevision` query to
     * speed up the network requests waterfall. To minimize the blast radius
     * of changes required to make it work, continue working with the `repo`
     * data as if it was received from a separate query.
     */
    const repoOrError = isErrorLike(resolvedRevisionOrError) ? resolvedRevisionOrError : resolvedRevisionOrError?.repo

    // The external links to show in the repository header, if any.
    const [externalLinks, setExternalLinks] = useState<ExternalLinkFields[] | undefined>()

    // The lifecycle props for repo header contributions.
    const [repoHeaderContributionsLifecycleProps, setRepoHeaderContributionsLifecycleProps] =
        useState<RepoHeaderContributionsLifecycleProps>()

    const childBreadcrumbSetters = props.useBreadcrumb(
        useMemo(() => {
            if (isErrorLike(resolvedRevisionOrError) || isErrorLike(repoOrError)) {
                return
            }

            const button = (
                <Button
                    to={resolvedRevisionOrError?.rootTreeURL || repoOrError?.url || ''}
                    disabled={!resolvedRevisionOrError}
                    className="text-nowrap test-repo-header-repo-link"
                    variant="secondary"
                    outline={true}
                    size="sm"
                    as={Link}
                >
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> {displayRepoName(repoName)}
                </Button>
            )

            return {
                key: 'repository',
                element: button,
            }
        }, [resolvedRevisionOrError, repoOrError, repoName])
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

        if (workspaceRootUri && extensionsController !== null) {
            extensionsController.extHostAPI
                .then(extensionHostAPI =>
                    extensionHostAPI.addWorkspaceRoot({
                        uri: workspaceRootUri,
                        inputRevision: revision || '',
                    })
                )
                .catch(error => {
                    logger.error('Error adding workspace root', error)
                })
        }

        // Clear the Sourcegraph extensions model's roots when navigating away.
        return () => {
            if (workspaceRootUri && extensionsController !== null) {
                extensionsController.extHostAPI
                    .then(extensionHostAPI => extensionHostAPI.removeWorkspaceRoot(workspaceRootUri))
                    .catch(error => {
                        logger.error('Error removing workspace root', error)
                    })
            }
        }
    }, [extensionsController, repoName, resolvedRevisionOrError, revision])

    // Update the navbar query to reflect the current repo / revision
    const [enableExperimentalQueryInput] = useExperimentalQueryInput()
    const queryPrefix = useMemo(
        () =>
            enableExperimentalQueryInput && selectedSearchContextSpec ? `context:${selectedSearchContextSpec} ` : '',
        [enableExperimentalQueryInput, selectedSearchContextSpec]
    )
    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)
    useEffect(() => {
        let query = queryPrefix + searchQueryForRepoRevision(repoName, revision)
        if (filePath) {
            query = `${query.trimEnd()} file:${escapeSpaces('^' + escapeRegExp(filePath))}`
        }
        onNavbarQueryChange({
            query,
            hint: EditorHint.Blur,
        })
    }, [revision, filePath, repoName, onNavbarQueryChange, queryPrefix])

    const isError = isErrorLike(repoOrError) || isErrorLike(resolvedRevisionOrError)

    // if revision for given repo does not resolve then we still proceed to render settings routes
    // while returning empty repository for all other routes
    const isEmptyRepo = isRevisionNotFoundErrorLike(repoOrError)

    // For repo errors beyond revision not found (aka empty repository)
    // we defer to RepoContainerError for every repo container request
    if (isError && !isEmptyRepo) {
        const viewerCanAdminister = !!authenticatedUser && authenticatedUser.siteAdmin

        return (
            <RepoContainerError
                repoName={repoName}
                viewerCanAdminister={viewerCanAdminister}
                repoFetchError={repoOrError as ErrorLike}
            />
        )
    }

    const repo = isError ? undefined : repoOrError
    const resolvedRevision = isError ? undefined : resolvedRevisionOrError
    const isBrainDotVisible = getIsBrainDotVisible({
        location,
        revision,
        repoName,
    })

    const repoRevisionContainerContext: RepoRevisionContainerContext = {
        ...props,
        ...repoHeaderContributionsLifecycleProps,
        ...childBreadcrumbSetters,
        repo,
        repoName,
        revision: revision || '',
        resolvedRevision,
    }

    const perforceCodeHostUrlToSwarmUrlMap =
        (props.settingsCascade.final &&
            !isErrorLike(props.settingsCascade.final) &&
            props.settingsCascade.final?.['perforce.codeHostToSwarmMap']) ||
        {}

    const repoContainerContext: Omit<RepoContainerContext, 'repo'> = {
        ...repoRevisionContainerContext,
        resolvedRevisionOrError,
        onDidUpdateExternalLinks: setExternalLinks,
        repoName,
    }

    return (
        <>
            {isCodyEnabled &&
                focusCodyShortcut?.keybindings.map((keybinding, index) => (
                    <Shortcut
                        key={index}
                        {...keybinding}
                        onMatch={() => {
                            setIsCodySidebarOpen(true)
                        }}
                    />
                ))}
            <div className={classNames('w-100 d-flex flex-row')}>
                <div className={classNames('w-100 d-flex flex-column', styles.repoContainer)}>
                    <RepoHeader
                        actionButtons={props.repoHeaderActionButtons}
                        breadcrumbs={props.breadcrumbs}
                        repoName={repoName}
                        revision={revision}
                        onLifecyclePropsChange={setRepoHeaderContributionsLifecycleProps}
                        settingsCascade={props.settingsCascade}
                        authenticatedUser={authenticatedUser}
                        platformContext={props.platformContext}
                        telemetryService={props.telemetryService}
                    />

                    {isCodyEnabled ? (
                        <RepoHeaderContributionPortal
                            position="right"
                            priority={1}
                            id="cody"
                            {...repoHeaderContributionsLifecycleProps}
                        >
                            {() =>
                                !isCodySidebarOpen ? (
                                    <AskCodyButton
                                        onClick={() => {
                                            props.telemetryService.log('web:codySidebar:chatOpened', {
                                                repo,
                                                path: filePath,
                                            })
                                            setIsCodySidebarOpen(true)
                                        }}
                                    />
                                ) : null
                            }
                        </RepoHeaderContributionPortal>
                    ) : null}

                    <RepoHeaderContributionPortal
                        position="right"
                        priority={2}
                        id="go-to-code-host"
                        {...repoHeaderContributionsLifecycleProps}
                    >
                        {({ actionType }) => (
                            <GoToCodeHostAction
                                repo={repo}
                                repoName={repoName}
                                // We need a revision to generate code host URLs, if revision isn't available, we use the default branch or HEAD.
                                revision={rawRevision || repo?.defaultBranch?.displayName || 'HEAD'}
                                filePath={filePath}
                                commitRange={commitRange}
                                range={range}
                                position={position}
                                perforceCodeHostUrlToSwarmUrlMap={perforceCodeHostUrlToSwarmUrlMap}
                                fetchFileExternalLinks={fetchFileExternalLinks}
                                actionType={actionType}
                                source="repoHeader"
                                key="go-to-code-host"
                                externalLinks={externalLinks}
                            />
                        )}
                    </RepoHeaderContributionPortal>

                    {isBrainDotVisible && (
                        <RepoHeaderContributionPortal
                            position="right"
                            priority={110}
                            id="code-intelligence-status"
                            {...repoHeaderContributionsLifecycleProps}
                        >
                            {({ actionType }) =>
                                props.brainDot && actionType === 'nav' ? (
                                    <props.brainDot
                                        key="code-intelligence-status"
                                        repoName={repoName}
                                        path={filePath}
                                        commit={resolvedRevision?.commitID ?? ''}
                                    />
                                ) : null
                            }
                        </RepoHeaderContributionPortal>
                    )}

                    <Suspense fallback={null}>
                        <Routes>
                            {repoContainerRoutes.map(({ path, render, condition = () => true }) => (
                                <Route
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    path={repoSplat + path}
                                    errorElement={<RouteError />}
                                    element={
                                        /**
                                         * `repoContainerRoutes` depend on `repo`. We render these routes only when
                                         * the `repo` value is resolved. If repo resolves to error due to empty repository
                                         * then we return Empty Repository.
                                         */
                                        repo && condition({ ...repoContainerContext, repo }) ? (
                                            render({ ...repoContainerContext, repo })
                                        ) : isEmptyRepo ? (
                                            <EmptyRepo />
                                        ) : null
                                    }
                                />
                            ))}
                            <Route
                                path={repoSplat + repoSettingsAreaPath}
                                errorElement={<RouteError />}
                                // Always render the `RepoSettingsArea` even for empty repo to allow side-admins access it.
                                element={<RepoSettingsArea {...repoRevisionContainerContext} repoName={repoName} />}
                            />
                            <Route
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                path="*"
                                errorElement={<RouteError />}
                                element={
                                    isEmptyRepo ? (
                                        <EmptyRepo />
                                    ) : (
                                        <RepoRevisionContainer
                                            {...repoRevisionContainerContext}
                                            {...childBreadcrumbSetters}
                                            routes={props.repoRevisionContainerRoutes}
                                        />
                                    )
                                }
                            />
                        </Routes>
                    </Suspense>
                </div>

                {isCodyEnabled && isCodySidebarOpen && (
                    <Panel
                        className="cody-sidebar-panel"
                        position="right"
                        ariaLabel="Cody sidebar"
                        maxSize={CODY_SIDEBAR_SIZES.max}
                        minSize={CODY_SIDEBAR_SIZES.min}
                        defaultSize={CODY_SIDEBAR_SIZES.default}
                        storageKey="size-cache-cody-sidebar"
                        onResize={onCodySidebarResize}
                    >
                        <CodyChat onClose={() => setIsCodySidebarOpen(false)} />
                    </Panel>
                )}
            </div>
        </>
    )
}

function getIsBrainDotVisible({
    location,
    repoName,
    revision,
}: {
    location: Location
    repoName: string
    revision: string | undefined
}): boolean {
    // Remove leading repository name and possible leading revision, then compare the remaining routes to
    // see if we should display the braindot badge for this route. We want this to be visible on the repo
    // root page, as well as directory and code views, but not administrative/non-code views.
    //
    // + 1 for the leading `/` in the pathname
    const matchRevisionAndRest = location.pathname.slice(repoName.length + 1)
    const matchOnlyRest =
        revision && matchRevisionAndRest.startsWith(`@${revision || ''}`)
            ? matchRevisionAndRest.slice(revision.length + 1)
            : matchRevisionAndRest

    return matchOnlyRest === '' || matchOnlyRest.startsWith('/-/tree') || matchOnlyRest.startsWith('/-/blob')
}

/**
 * Performs a redirect to the host of the given URL with the path, query etc. properties of the current URL.
 */
function redirectToExternalHost(externalRedirectURL: string): void {
    const externalHostURL = new URL(externalRedirectURL)
    const redirectURL = new URL(window.location.href)
    // Preserve the path of the current URL and redirect to the repo on the external host.
    redirectURL.host = externalHostURL.host
    redirectURL.protocol = externalHostURL.protocol
    window.location.replace(redirectURL.href)
}

const EmptyRepo: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={RepoQuestionIcon} title="Empty repository" />
)
