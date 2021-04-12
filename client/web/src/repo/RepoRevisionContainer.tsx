import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import React, { useMemo } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { UncontrolledPopover } from 'reactstrap'

import {
    CloneInProgressError,
    isCloneInProgressErrorLike,
    isRevisionNotFoundErrorLike,
    isRepoNotFoundErrorLike,
} from '@sourcegraph/shared/src/backend/errors'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { AuthenticatedUser } from '../auth'
import { ErrorMessage } from '../components/alerts'
import { BreadcrumbSetters } from '../components/Breadcrumbs'
import { HeroPage } from '../components/HeroPage'
import { ActionItemsBarProps } from '../extensions/components/ActionItemsBar'
import { RepositoryFields } from '../graphql-operations'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps, SearchContextProps } from '../search'
import { RouteDescriptor } from '../util/contributions'

import { CopyLinkAction } from './actions/CopyLinkAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { ResolvedRevision } from './backend'
import { HoverThresholdProps, RepoContainerContext } from './RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { EmptyRepositoryPage, RepositoryCloningInProgressPage } from './RepositoryGitDataContainer'
import { RevisionsPopover } from './RevisionsPopover'
import { RepoSettingsAreaRoute } from './settings/RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './settings/RepoSettingsSidebar'

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
        Omit<RepoContainerContext, 'onDidUpdateExternalLinks'>,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps {
    repo: RepositoryFields
    resolvedRev: ResolvedRevision

    /** The URL route match for {@link RepoRevisionContainer}. */
    routePrefix: string

    globbing: boolean
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
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        RevisionSpec,
        BreadcrumbSetters,
        ActionItemsBarProps {
    routes: readonly RepoRevisionContainerRoute[]
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
    routePrefix: string

    /**
     * The resolved revision or an error if it could not be resolved. This value lives in RepoContainer (this
     * component's parent) but originates from this component.
     */
    resolvedRevisionOrError: ResolvedRevision | ErrorLike | undefined

    history: H.History

    globbing: boolean
}

/**
 * A container for a repository page that incorporates revisioned Git data. (For example,
 * blob and tree pages are revisioned, but the repository settings page is not.)
 */
export const RepoRevisionContainer: React.FunctionComponent<RepoRevisionContainerProps> = ({
    useBreadcrumb,
    ...props
}) => {
    const breadcrumbSetters = useBreadcrumb(
        useMemo(() => {
            if (!props.resolvedRevisionOrError || isErrorLike(props.resolvedRevisionOrError)) {
                return
            }

            return {
                key: 'revision',
                divider: <span className="mr-1">@</span>,
                element: (
                    <div className="d-flex align-items-center" key="repo-revision">
                        <span className="test-revision">
                            {(props.revision && props.revision === props.resolvedRevisionOrError.commitID
                                ? props.resolvedRevisionOrError.commitID.slice(0, 7)
                                : props.revision) ||
                                props.resolvedRevisionOrError.defaultBranch ||
                                'HEAD'}
                        </span>
                        <button
                            type="button"
                            id="repo-revision-popover"
                            className="btn btn-icon px-0"
                            aria-label="Change revision"
                        >
                            <MenuDownIcon className="icon-inline" />
                        </button>
                        <UncontrolledPopover
                            placement="bottom-start"
                            target="repo-revision-popover"
                            trigger="legacy"
                            hideArrow={true}
                            popperClassName="border-0"
                        >
                            <RevisionsPopover
                                repo={props.repo.id}
                                repoName={props.repo.name}
                                defaultBranch={props.resolvedRevisionOrError.defaultBranch}
                                currentRev={props.revision}
                                currentCommitID={props.resolvedRevisionOrError.commitID}
                                history={props.history}
                                location={props.location}
                            />
                        </UncontrolledPopover>
                    </div>
                ),
            }
        }, [
            props.revision,
            props.resolvedRevisionOrError,
            props.repo.id,
            props.repo.name,
            props.history,
            props.location,
        ])
    )

    if (!props.resolvedRevisionOrError) {
        // Render nothing while loading
        return null
    }

    if (isErrorLike(props.resolvedRevisionOrError)) {
        // Show error page
        if (isCloneInProgressErrorLike(props.resolvedRevisionOrError)) {
            return (
                <RepositoryCloningInProgressPage
                    repoName={props.repo.name}
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

    const context: RepoRevisionContainerContext = {
        ...props,
        ...breadcrumbSetters,
        resolvedRev: props.resolvedRevisionOrError,
    }

    const resolvedRevisionOrError = props.resolvedRevisionOrError

    return (
        <div className="repo-revision-container">
            <Switch>
                {/* eslint-disable react/jsx-no-bind */}
                {props.routes.map(
                    ({ path, render, exact, condition = () => true }) =>
                        condition(context) && (
                            <Route
                                path={props.routePrefix + path}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={exact}
                                render={routeComponentProps => render({ ...context, ...routeComponentProps })}
                            />
                        )
                )}
            </Switch>
            <RepoHeaderContributionPortal
                position="left"
                id="copy-link"
                repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
            >
                {() => <CopyLinkAction key="copy-link" location={props.location} />}
            </RepoHeaderContributionPortal>
            <RepoHeaderContributionPortal
                position="right"
                priority={3}
                id="go-to-permalink"
                repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
            >
                {context => (
                    <GoToPermalinkAction
                        key="go-to-permalink"
                        revision={props.revision}
                        commitID={resolvedRevisionOrError.commitID}
                        location={props.location}
                        history={props.history}
                        {...context}
                    />
                )}
            </RepoHeaderContributionPortal>
        </div>
    )
}
