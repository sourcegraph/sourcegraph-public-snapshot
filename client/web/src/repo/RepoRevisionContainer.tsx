import React, { useCallback, useMemo, useState } from 'react'

import * as H from 'history'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { StreamingSearchResultsListProps, CopyPathAction } from '@sourcegraph/branded'
import { isErrorLike } from '@sourcegraph/common'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button, LoadingSpinner, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { CodeIntelligenceProps } from '../codeintel'
import { BreadcrumbSetters } from '../components/Breadcrumbs'
import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'
import { RepositoryFields } from '../graphql-operations'
import { CodeInsightsProps } from '../insights/types'
import { NotebookProps } from '../notebooks'
import { SearchStreamingProps } from '../search'
import { eventLogger } from '../tracking/eventLogger'
import { RouteDescriptor } from '../util/contributions'
import { parseBrowserRepoURL } from '../util/url'

import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ResolvedRevision } from './backend'
import { RepoRevisionChevronDownIcon, RepoRevisionWrapper } from './components/RepoRevision'
import { HoverThresholdProps, RepoContainerContext } from './RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RevisionsPopover } from './RevisionsPopover'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './settings/RepoSettingsSidebar'

import styles from './RepoRevisionContainer.module.scss'

/** Props passed to sub-routes of {@link RepoRevisionContainer}. */
export interface RepoRevisionContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        HoverThresholdProps,
        Omit<RepoContainerContext, 'onDidUpdateExternalLinks' | 'repo' | 'resolvedRevisionOrError'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        BatchChangesProps,
        Pick<CodeIntelligenceProps, 'codeIntelligenceEnabled' | 'useCodeIntel'>,
        CodeInsightsProps,
        NotebookProps {
    repo: RepositoryFields | undefined
    resolvedRevision: ResolvedRevision | undefined

    repoName: string

    /** The URL route match for {@link RepoRevisionContainer}. */
    routePrefix: string

    globbing: boolean

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

/** A sub-route of {@link RepoRevisionContainer}. */
export interface RepoRevisionContainerRoute extends RouteDescriptor<RepoRevisionContainerContext> {}

interface RepoRevisionContainerProps
    extends RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        NotebookProps {
    routes: readonly RepoRevisionContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    repo: RepositoryFields | undefined
    authenticatedUser: AuthenticatedUser | null
    routePrefix: string

    /**
     * The resolved revision or an error if it could not be resolved.
     */
    resolvedRevision: ResolvedRevision | undefined

    /** The repoName from the URL */
    repoName: string

    history: H.History

    globbing: boolean

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

interface RepoRevisionBreadcrumbProps extends Pick<RepoRevisionContainerProps, 'repo' | 'revision' | 'repoName'> {
    resolvedRevision: ResolvedRevision | undefined
}

export const RepoRevisionContainerBreadcrumb: React.FunctionComponent<
    React.PropsWithChildren<RepoRevisionBreadcrumbProps>
> = ({ revision, resolvedRevision, repoName, repo }) => {
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
export const RepoRevisionContainer: React.FunctionComponent<React.PropsWithChildren<RepoRevisionContainerProps>> = ({
    useBreadcrumb,
    ...props
}) => {
    const breadcrumbSetters = useBreadcrumb(
        useMemo(() => {
            if (isErrorLike(props.resolvedRevision)) {
                return
            }

            return {
                key: 'revision',
                divider: <span className={styles.divider}>@</span>,
                element: (
                    <RepoRevisionContainerBreadcrumb
                        resolvedRevision={props.resolvedRevision}
                        revision={props.revision}
                        repoName={props.repoName}
                        repo={props.repo}
                    />
                ),
            }
        }, [props.resolvedRevision, props.revision, props.repo, props.repoName])
    )

    const repoRevisionContainerContext: RepoRevisionContainerContext = {
        ...props,
        ...breadcrumbSetters,
        resolvedRevision: props.resolvedRevision,
    }

    const resolvedRevision = props.resolvedRevision

    const { repoName, filePath } = parseBrowserRepoURL(location.pathname)

    return (
        <>
            <RepoRevisionWrapper className="pl-3">
                <Switch>
                    {props.routes.map(
                        ({ path, render, exact, condition = () => true }) =>
                            condition(repoRevisionContainerContext) && (
                                <Route
                                    path={props.routePrefix + path}
                                    key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                    exact={exact}
                                    render={routeComponentProps =>
                                        render({
                                            ...repoRevisionContainerContext,
                                            ...routeComponentProps,
                                        })
                                    }
                                />
                            )
                    )}
                </Switch>
                <RepoHeaderContributionPortal
                    position="left"
                    id="copy-path"
                    repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
                >
                    {() => (
                        <CopyPathAction
                            telemetryService={eventLogger}
                            filePath={filePath || repoName}
                            key="copy-path"
                        />
                    )}
                </RepoHeaderContributionPortal>
                {resolvedRevision && (
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
                                revision={props.revision}
                                commitID={resolvedRevision.commitID}
                                location={props.location}
                                history={props.history}
                                {...context}
                            />
                        )}
                    </RepoHeaderContributionPortal>
                )}
            </RepoRevisionWrapper>
        </>
    )
}
