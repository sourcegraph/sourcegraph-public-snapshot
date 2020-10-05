import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import TagIcon from 'mdi-react/TagIcon'
import UserIcon from 'mdi-react/UserIcon'
import React, { useMemo, useEffect } from 'react'
import { Link, Redirect } from 'react-router-dom'
import { EMPTY } from 'rxjs'
import { catchError } from 'rxjs/operators'
import { ActionItem } from '../../../../shared/src/actions/ActionItem'
import { ActionsContainer } from '../../../../shared/src/actions/ActionsContainer'
import { ContributableMenu, ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { displayRepoName } from '../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { PageTitle } from '../../components/PageTitle'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps } from '../../search'
import { basename } from '../../util/path'
import { fetchTreeEntries } from '../backend'
import { ThemeProps } from '../../../../shared/src/theme'
import { ErrorAlert } from '../../components/alerts'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { toPrettyBlobURL, toURIWithPath } from '../../../../shared/src/util/url'
import { getViewsForContainer } from '../../../../shared/src/api/client/services/viewService'
import { Settings } from '../../schema/settings.schema'
import { ViewGrid } from './ViewGrid'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { SearchInThisGraphButton } from '../../enterprise/graphs/search/SearchInThisGraphButton'
import { GraphSelectionProps, SelectableGraph } from '../../enterprise/graphs/selector/graphSelectionProps'
import { escapeRegExp } from 'lodash'
import PackageVariantClosedIcon from 'mdi-react/PackageVariantClosedIcon'

interface Props
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        GraphSelectionProps,
        BreadcrumbSetters {
    repoName: string
    repoID: GQL.ID
    repoDescription: string
    /** The tree's path in TreePage. We call it filePath for consistency elsewhere. */
    filePath: string
    commitID: string
    revision: string
    location: H.Location
    history: H.History
    globbing: boolean

    /**
     * The contextual graph that consists of this repository.
     */
    repositoryContextualGraph: SelectableGraph
}

export const TreePage: React.FunctionComponent<Props> = ({
    repoName,
    repoID,
    repoDescription,
    commitID,
    revision,
    filePath,
    patternType,
    caseSensitive,
    settingsCascade,
    useBreadcrumb,
    repositoryContextualGraph,
    ...props
}) => {
    useEffect(() => {
        if (filePath === '') {
            props.telemetryService.logViewEvent('Repository')
        } else {
            props.telemetryService.logViewEvent('Tree')
        }
    }, [filePath, props.telemetryService])

    useBreadcrumb(
        useMemo(() => {
            if (!filePath) {
                return
            }
            return {
                key: 'treePath',
                className: 'flex-shrink-past-contents',
                element: (
                    <FilePathBreadcrumbs
                        key="path"
                        repoName={repoName}
                        revision={revision}
                        filePath={filePath}
                        isDir={true}
                    />
                ),
            }
        }, [repoName, revision, filePath])
    )

    const treeOrError = useObservable(
        useMemo(
            () =>
                fetchTreeEntries({
                    repoName,
                    commitID,
                    revision,
                    filePath,
                    first: 2500,
                }).pipe(catchError((error): [ErrorLike] => [asError(error)])),
            [repoName, commitID, revision, filePath]
        )
    )

    const { services } = props.extensionsController

    const codeInsightsEnabled =
        !isErrorLike(settingsCascade.final) && !!settingsCascade.final?.experimentalFeatures?.codeInsights

    // Add DirectoryViewer
    const uri = toURIWithPath({ repoName, commitID, filePath })
    useEffect(() => {
        if (!codeInsightsEnabled) {
            return
        }
        const viewerId = services.viewer.addViewer({
            type: 'DirectoryViewer',
            isActive: true,
            resource: uri,
        })
        return () => services.viewer.removeViewer(viewerId)
    }, [services.viewer, services.model, uri, codeInsightsEnabled])

    // Observe directory views
    const workspaceUri = services.workspace.roots.value[0]?.uri
    const views = useObservable(
        useMemo(
            () =>
                codeInsightsEnabled && workspaceUri
                    ? getViewsForContainer(
                          ContributableViewContainer.Directory,
                          {
                              viewer: {
                                  type: 'DirectoryViewer',
                                  directory: {
                                      uri,
                                  },
                              },
                              workspace: {
                                  uri: workspaceUri,
                              },
                          },
                          services.view
                      )
                    : EMPTY,
            [codeInsightsEnabled, workspaceUri, uri, services.view]
        )
    )

    const getPageTitle = (): string => {
        const repoString = displayRepoName(repoName)
        if (filePath) {
            return `${basename(filePath)} - ${repoString}`
        }
        return `${repoString}`
    }

    return (
        <div className="tree-page">
            <PageTitle title={getPageTitle()} />
            {treeOrError === undefined ? (
                <div>
                    <LoadingSpinner className="icon-inline tree-page__entries-loader" /> Loading files and directories
                </div>
            ) : isErrorLike(treeOrError) ? (
                // If the tree is actually a blob, be helpful and redirect to the blob page.
                // We don't have error names on GraphQL errors.
                /not a directory/i.test(treeOrError.message) ? (
                    <Redirect to={toPrettyBlobURL({ repoName, revision, commitID, filePath })} />
                ) : (
                    <ErrorAlert error={treeOrError} history={props.history} />
                )
            ) : (
                <>
                    <header className="mb-3">
                        {treeOrError.isRoot && repoDescription && <p className="mb-0">{repoDescription}</p>}
                        <SearchInThisGraphButton
                            {...props}
                            graph={repositoryContextualGraph}
                            // TODO(sqs): escapeRegExp doesn't work with paths with spaces
                            query={treeOrError.isRoot ? undefined : `file:^${escapeRegExp(filePath)}/`}
                            className="mt-3 d-block"
                        >
                            Search in this {treeOrError.isRoot ? 'repository' : 'directory'}
                        </SearchInThisGraphButton>
                        {treeOrError.isRoot && (
                            <>
                                <div>
                                    <div className="btn-group mt-2 d-none">
                                        {/* TODO(sqs) */}
                                        <Link className="btn btn-secondary" to={`${treeOrError.url}/-/commits`}>
                                            <SourceCommitIcon className="icon-inline" /> 173 commits
                                        </Link>
                                        <SearchInThisGraphButton
                                            {...props}
                                            graph={repositoryContextualGraph}
                                            // TODO(sqs): escapeRegExp doesn't work with paths with spaces
                                            query={treeOrError.isRoot ? undefined : `file:^${escapeRegExp(filePath)}/`}
                                        >
                                            Search in diffs &amp; messages
                                        </SearchInThisGraphButton>
                                    </div>
                                </div>
                                <div>
                                    <div className="btn-group mt-2 d-none">
                                        {' '}
                                        {/* TODO(sqs) */}
                                        <Link className="btn btn-secondary" to={`/${repoName}/-/branches`}>
                                            <SourceBranchIcon className="icon-inline" /> 35 branches
                                        </Link>
                                        <Link className="btn btn-secondary" to={`/${repoName}/-/tags`}>
                                            <TagIcon className="icon-inline" /> 17 tags
                                        </Link>
                                        <SearchInThisGraphButton
                                            {...props}
                                            graph={repositoryContextualGraph}
                                            // TODO(sqs): escapeRegExp doesn't work with paths with spaces
                                            query={treeOrError.isRoot ? undefined : `file:^${escapeRegExp(filePath)}/`}
                                        >
                                            Search in multiple branches/tags
                                        </SearchInThisGraphButton>
                                    </div>
                                </div>
                                <div>
                                    <div className="btn-group mt-2 d-none">
                                        <Link className="btn btn-secondary" to={`/${repoName}/-/packages`}>
                                            <PackageVariantClosedIcon className="icon-inline" /> 17 packages
                                        </Link>
                                        <button type="button" className="btn btn-outline-secondary">
                                            Go to package...
                                        </button>
                                    </div>
                                </div>
                                <div>
                                    <div className="btn-group mt-2 d-none">
                                        <Link className="btn btn-secondary" to={`/${repoName}/-/packages`}>
                                            <SourceBranchIcon className="icon-inline" /> 51 dependencies
                                        </Link>
                                        <SearchInThisGraphButton
                                            {...props}
                                            graph={repositoryContextualGraph}
                                            // TODO(sqs): escapeRegExp doesn't work with paths with spaces
                                            query="TODO"
                                        >
                                            Search in dependencies
                                        </SearchInThisGraphButton>
                                    </div>
                                </div>
                                <div>
                                    <div className="btn-group mt-2 d-none">
                                        <Link
                                            className="btn btn-secondary"
                                            to={
                                                revision
                                                    ? `/${repoName}/-/compare/...${encodeURIComponent(revision)}`
                                                    : `/${repoName}/-/compare`
                                            }
                                        >
                                            <HistoryIcon className="icon-inline" /> Compare
                                        </Link>
                                        <Link className="btn btn-secondary" to={`/${repoName}/-/stats/contributors`}>
                                            <UserIcon className="icon-inline" /> Contributors
                                        </Link>
                                    </div>
                                </div>
                            </>
                        )}
                    </header>
                    {views && (
                        <ViewGrid
                            {...props}
                            className="tree-page__section mb-5"
                            viewGridStorageKey="tree-page"
                            views={views}
                            patternType={patternType}
                            settingsCascade={settingsCascade}
                            caseSensitive={caseSensitive}
                        />
                    )}
                    {/* eslint-disable react/jsx-no-bind */}
                    <ActionsContainer
                        {...props}
                        menu={ContributableMenu.DirectoryPage}
                        render={items => (
                            <section className="tree-page__section">
                                <h3 className="tree-page__section-header">Actions</h3>
                                {items.map(item => (
                                    <ActionItem
                                        {...props}
                                        key={item.action.id}
                                        {...item}
                                        className="btn btn-secondary mr-1 mb-1"
                                    />
                                ))}
                            </section>
                        )}
                        empty={null}
                    />
                </>
            )}
        </div>
    )
}
