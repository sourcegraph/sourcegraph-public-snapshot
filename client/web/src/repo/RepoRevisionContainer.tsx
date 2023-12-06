import { type FC, useCallback, useMemo, useState } from 'react'

import { Route, Routes } from 'react-router-dom'

import type { StreamingSearchResultsListProps } from '@sourcegraph/branded'
import { isErrorLike } from '@sourcegraph/common'
import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SearchContextProps } from '@sourcegraph/shared/src/search'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import type { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button, LoadingSpinner, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import type { BatchChangesProps } from '../batches'
import type { CodeIntelligenceProps } from '../codeintel'
import type { BreadcrumbSetters } from '../components/Breadcrumbs'
import type { RepositoryFields } from '../graphql-operations'
import type { CodeInsightsProps } from '../insights/types'
import type { NotebookProps } from '../notebooks'
import type { OwnConfigProps } from '../own/OwnConfigProps'
import type { SearchStreamingProps } from '../search'
import type { RouteV6Descriptor } from '../util/contributions'

import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import type { ResolvedRevision } from './backend'
import { RepoRevisionChevronDownIcon, RepoRevisionWrapper } from './components/RepoRevision'
import { isPackageServiceType } from './packages/isPackageServiceType'
import type { HoverThresholdProps, RepoContainerContext } from './RepoContainer'
import type { RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RevisionsPopover } from './RevisionsPopover'
import type { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import type { RepoSettingsSideBarGroup } from './settings/RepoSettingsSidebar'

import styles from './RepoRevisionContainer.module.scss'

/** Props passed to sub-routes of {@link RepoRevisionContainer}. */
export interface RepoRevisionContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        TelemetryProps,
        TelemetryV2Props,
        HoverThresholdProps,
        Omit<RepoContainerContext, 'onDidUpdateExternalLinks' | 'repo' | 'resolvedRevisionOrError'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        RevisionSpec,
        BreadcrumbSetters,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        BatchChangesProps,
        Pick<CodeIntelligenceProps, 'codeIntelligenceEnabled' | 'useCodeIntel'>,
        CodeInsightsProps,
        NotebookProps,
        OwnConfigProps {
    repo: RepositoryFields | undefined
    resolvedRevision: ResolvedRevision | undefined

    repoName: string

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

/** A sub-route of {@link RepoRevisionContainer}. */
export interface RepoRevisionContainerRoute extends RouteV6Descriptor<RepoRevisionContainerContext> {}

interface RepoRevisionContainerProps
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        TelemetryV2Props,
        HoverThresholdProps,
        ExtensionsControllerProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        RevisionSpec,
        BreadcrumbSetters,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        NotebookProps,
        OwnConfigProps {
    routes: readonly RepoRevisionContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    repo: RepositoryFields | undefined
    authenticatedUser: AuthenticatedUser | null

    /**
     * The resolved revision or an error if it could not be resolved.
     */
    resolvedRevision: ResolvedRevision | undefined

    /** The repoName from the URL */
    repoName: string

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

interface RepoRevisionBreadcrumbProps extends Pick<RepoRevisionContainerProps, 'repo' | 'revision' | 'repoName'> {
    resolvedRevision: ResolvedRevision | undefined
}

export const RepoRevisionContainerBreadcrumb: FC<RepoRevisionBreadcrumbProps> = props => {
    const { revision, resolvedRevision, repoName, repo } = props

    const [popoverOpen, setPopoverOpen] = useState(false)
    const togglePopover = useCallback(() => setPopoverOpen(previous => !previous), [])

    const revisionLabel = (revision && revision === resolvedRevision?.commitID
        ? resolvedRevision?.commitID.slice(0, 7)
        : revision.slice(0, 7)) ||
        resolvedRevision?.defaultBranch || <LoadingSpinner />

    const isPopoverContentReady = repo && resolvedRevision

    return (
        <Popover isOpen={popoverOpen} onOpenChange={event => setPopoverOpen(event.isOpen)}>
            <PopoverTrigger
                as={Button}
                className="d-flex align-items-center text-nowrap"
                key="repo-revision"
                id="repo-revision-popover"
                aria-label="Change revision"
                outline={true}
                variant="secondary"
                size="sm"
                disabled={!isPopoverContentReady}
            >
                {revisionLabel}
                <RepoRevisionChevronDownIcon aria-hidden={true} />
            </PopoverTrigger>
            <PopoverContent
                position={Position.bottomStart}
                className="pt-0 pb-0"
                aria-labelledby="repo-revision-popover"
            >
                {isPopoverContentReady && (
                    <RevisionsPopover
                        repoId={repo?.id}
                        repoName={repoName}
                        repoServiceType={repo?.externalRepository?.serviceType}
                        defaultBranch={resolvedRevision?.defaultBranch}
                        currentRev={revision}
                        currentCommitID={resolvedRevision?.commitID}
                        togglePopover={togglePopover}
                        onSelect={togglePopover}
                    />
                )}
            </PopoverContent>
        </Popover>
    )
}

/**
 * A container for a repository page that incorporates revisioned Git data. (For example,
 * blob and tree pages are revisioned, but the repository settings page is not.)
 */
export const RepoRevisionContainer: FC<RepoRevisionContainerProps> = props => {
    const { useBreadcrumb, resolvedRevision, revision, repo, repoName, routes } = props

    const breadcrumbSetters = useBreadcrumb(
        useMemo(() => {
            if (isErrorLike(resolvedRevision)) {
                return
            }

            return {
                key: 'revision',
                divider: <span className={styles.divider}>@</span>,
                element: (
                    <RepoRevisionContainerBreadcrumb
                        resolvedRevision={resolvedRevision}
                        revision={revision}
                        repoName={repoName}
                        repo={repo}
                    />
                ),
            }
        }, [resolvedRevision, revision, repo, repoName])
    )

    const isPackage = useMemo(
        () => isPackageServiceType(repo?.externalRepository.serviceType),
        [repo?.externalRepository.serviceType]
    )

    const repoRevisionContainerContext: RepoRevisionContainerContext = {
        ...props,
        ...breadcrumbSetters,
        resolvedRevision,
    }

    return (
        <RepoRevisionWrapper className="px-3">
            <Routes>
                {routes.map(
                    ({ path, render, condition = () => true }) =>
                        condition(repoRevisionContainerContext) && (
                            <Route key="hardcoded-key" path={path} element={render(repoRevisionContainerContext)} />
                        )
                )}
            </Routes>
            {resolvedRevision && !isPackage && (
                <RepoHeaderContributionPortal
                    position="right"
                    priority={3}
                    id="go-to-permalink"
                    repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
                >
                    {context => (
                        <GoToPermalinkAction
                            key="go-to-permalink"
                            telemetryService={props.telemetryService}
                            telemetryRecorder={props.telemetryRecorder}
                            revision={props.revision}
                            commitID={resolvedRevision.commitID}
                            {...context}
                        />
                    )}
                </RepoHeaderContributionPortal>
            )}
        </RepoRevisionWrapper>
    )
}
