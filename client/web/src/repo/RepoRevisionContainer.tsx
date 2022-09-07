import React, { useCallback, useMemo, useState } from 'react'

import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import {
    CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
    isRepoNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button, LoadingSpinner, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { CodeIntelligenceProps } from '../codeintel'
import { BreadcrumbSetters } from '../components/Breadcrumbs'
import { HeroPage } from '../components/HeroPage'
import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'
import { RepositoryFields } from '../graphql-operations'
import { CodeInsightsProps } from '../insights/types'
import { SearchStreamingProps } from '../search'
import { RouteDescriptor } from '../util/contributions'

import { CopyPathAction } from './actions/CopyPathAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ResolvedRevision } from './backend'
import { BlobProps } from './blob/Blob'
import { RepoRevisionChevronDownIcon, RepoRevisionWrapper } from './components/RepoRevision'
import { HoverThresholdProps, RepoContainerContext } from './RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { EmptyRepositoryPage, RepositoryCloningInProgressPage } from './RepositoryGitDataContainer'
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
        ActivationProps,
        Omit<RepoContainerContext, 'onDidUpdateExternalLinks' | 'repo'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        Pick<BlobProps, 'onHandleFuzzyFinder'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        BatchChangesProps,
        CodeInsightsProps {
    repo: RepositoryFields | undefined
    resolvedRev: ResolvedRevision | undefined

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
        ActivationProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        Pick<BlobProps, 'onHandleFuzzyFinder'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps {
    routes: readonly RepoRevisionContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    repo: RepositoryFields | undefined
    authenticatedUser: AuthenticatedUser | null
    routePrefix: string

    /**
     * The resolved revision or an error if it could not be resolved.
     */
    resolvedRevisionOrError: ResolvedRevision | ErrorLike | undefined

    /** The repoName from the URL */
    repoName: string

    history: H.History

    globbing: boolean

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}

interface RepoRevisionBreadcrumbProps extends Pick<RepoRevisionContainerProps, 'repo' | 'revision' | 'repoName'> {
    resolvedRevisionOrError: ResolvedRevision | undefined
}

const RepoRevisionContainerBreadcrumb: React.FunctionComponent<
    React.PropsWithChildren<RepoRevisionBreadcrumbProps>
> = ({ revision, resolvedRevisionOrError, repoName, repo }) => {
    const [popoverOpen, setPopoverOpen] = useState(false)
    const togglePopover = useCallback(() => setPopoverOpen(previous => !previous), [])

    const revisionLabel = (revision && revision === resolvedRevisionOrError?.commitID
        ? resolvedRevisionOrError?.commitID.slice(0, 7)
        : revision.slice(0, 7)) ||
        resolvedRevisionOrError?.defaultBranch || <LoadingSpinner />

    const isPopoverContentReady = repo && resolvedRevisionOrError

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
                        defaultBranch={resolvedRevisionOrError?.defaultBranch}
                        currentRev={revision}
                        currentCommitID={resolvedRevisionOrError?.commitID}
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
            if (isErrorLike(props.resolvedRevisionOrError)) {
                return
            }

            return {
                key: 'revision',
                divider: <span className={styles.divider}>@</span>,
                element: (
                    <RepoRevisionContainerBreadcrumb
                        resolvedRevisionOrError={props.resolvedRevisionOrError}
                        revision={props.revision}
                        repoName={props.repoName}
                        repo={props.repo}
                    />
                ),
            }
        }, [props.resolvedRevisionOrError, props.revision, props.repo, props.repoName])
    )

    if (isErrorLike(props.resolvedRevisionOrError)) {
        // Show error page
        if (isCloneInProgressErrorLike(props.resolvedRevisionOrError)) {
            return (
                <RepositoryCloningInProgressPage
                    repoName={props.repoName}
                    progress={(props.resolvedRevisionOrError as CloneInProgressError).progress}
                />
            )
        }
        if (isRepoNotFoundErrorLike(props.resolvedRevisionOrError)) {
            return (
                <HeroPage
                    icon={MapSearchIcon}
                    title="404: Not Found"
                    subtitle="The requested repository was not found."
                />
            )
        }
        if (isRevisionNotFoundErrorLike(props.resolvedRevisionOrError)) {
            if (!props.revision) {
                return <EmptyRepositoryPage />
            }
            return (
                <HeroPage
                    icon={MapSearchIcon}
                    title="404: Not Found"
                    subtitle="The requested revision was not found."
                />
            )
        }
        return (
            <HeroPage
                icon={AlertCircleIcon}
                title="Error"
                subtitle={<ErrorMessage error={props.resolvedRevisionOrError} />}
            />
        )
    }

    const repoRevisionContainerContext: RepoRevisionContainerContext = {
        ...props,
        ...breadcrumbSetters,
        resolvedRev: props.resolvedRevisionOrError,
    }

    const resolvedRevisionOrError = props.resolvedRevisionOrError

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
                    {() => <CopyPathAction key="copy-path" />}
                </RepoHeaderContributionPortal>
                {resolvedRevisionOrError && (
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
                                commitID={resolvedRevisionOrError.commitID}
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
