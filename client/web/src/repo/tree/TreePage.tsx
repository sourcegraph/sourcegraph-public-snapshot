import classNames from 'classnames'
import * as H from 'history'
import FolderIcon from 'mdi-react/FolderIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useMemo, useEffect } from 'react'
import { Redirect } from 'react-router-dom'
import { EMPTY } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { getFileDecorations } from '../../backend/features'
import { BatchChangesProps } from '../../batches'
import { CodeIntelligenceProps } from '../../codeintel'
import { ErrorAlert } from '../../components/alerts'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { TreePageRepositoryFields } from '../../graphql-operations'
import { CodeInsightsProps } from '../../insights/types'
import { Settings } from '../../schema/settings.schema'
import { PatternTypeProps, SearchContextProps } from '../../search'
import { basename } from '../../util/path'
import { fetchTreeEntries } from '../backend'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { RepoSidebarViewOptionsProps } from '../RepoRevisionSidebar'

import { TreeCommits } from './commits/TreeCommits'
import { RepositoryRootLinks } from './RepositoryRootLinks'
import { TreeEntriesSection } from './TreeEntriesSection'
import styles from './TreePage.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        PatternTypeProps,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        RepoSidebarViewOptionsProps,
        BreadcrumbSetters {
    repo: TreePageRepositoryFields
    /** The tree's path in TreePage. We call it filePath for consistency elsewhere. */
    filePath: string
    commitID: string
    revision: string
    location: H.Location
    history: H.History
    globbing: boolean
}

export const treePageRepositoryFragment = gql`
    fragment TreePageRepositoryFields on Repository {
        id
        name
        description
        viewerCanAdminister
        url
    }
`

export const TreePage: React.FunctionComponent<Props> = ({
    repo,
    commitID,
    revision,
    filePath,
    patternType,
    settingsCascade,
    useBreadcrumb,
    codeIntelligenceEnabled,
    batchChangesEnabled,
    extensionViews: ExtensionViewsSection,
    repoSidebarIsVisible,
    setRepoSidebarIsVisible,
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
                        repoName={repo.name}
                        revision={revision}
                        filePath={filePath}
                        isDir={true}
                        repoUrl={repo.url}
                        telemetryService={props.telemetryService}
                    />
                ),
            }
        }, [repo.name, repo.url, revision, filePath, props.telemetryService])
    )

    const treeOrError = useObservable(
        useMemo(
            () =>
                fetchTreeEntries({
                    repoName: repo.name,
                    commitID,
                    revision,
                    filePath,
                    first: 2500,
                }).pipe(catchError((error): [ErrorLike] => [asError(error)])),
            [repo.name, commitID, revision, filePath]
        )
    )

    const fileDecorationsByPath =
        useObservable<FileDecorationsByPath>(
            useMemo(
                () =>
                    treeOrError && !isErrorLike(treeOrError)
                        ? getFileDecorations({
                              files: treeOrError.entries,
                              extensionsController: props.extensionsController,
                              repoName: repo.name,
                              commitID,
                              parentNodeUri: treeOrError.url,
                          })
                        : EMPTY,
                [treeOrError, repo.name, commitID, props.extensionsController]
            )
        ) ?? {}

    const showCodeInsights =
        !isErrorLike(settingsCascade.final) &&
        !!settingsCascade.final?.experimentalFeatures?.codeInsights &&
        settingsCascade.final['insights.displayLocation.directory'] === true

    // Add DirectoryViewer
    const uri = toURIWithPath({ repoName: repo.name, commitID, filePath })

    useEffect(() => {
        if (!showCodeInsights) {
            return
        }

        const viewerIdPromise = props.extensionsController.extHostAPI
            .then(extensionHostAPI =>
                extensionHostAPI.addViewerIfNotExists({
                    type: 'DirectoryViewer',
                    isActive: true,
                    resource: uri,
                })
            )
            .catch(error => {
                console.error('Error adding viewer to extension host:', error)
                return null
            })

        return () => {
            Promise.all([props.extensionsController.extHostAPI, viewerIdPromise])
                .then(([extensionHostAPI, viewerId]) => {
                    if (viewerId) {
                        return extensionHostAPI.removeViewer(viewerId)
                    }
                    return
                })
                .catch(error => console.error('Error removing viewer from extension host:', error))
        }
    }, [uri, showCodeInsights, props.extensionsController])

    const getPageTitle = (): string => {
        const repoString = displayRepoName(repo.name)
        if (filePath) {
            return `${basename(filePath)} - ${repoString}`
        }
        return `${repoString}`
    }

    return (
        <div className={styles.treePage}>
            <Container className={styles.container}>
                <PageTitle title={getPageTitle()} />
                {treeOrError === undefined ? (
                    <div>
                        <LoadingSpinner className="icon-inline" /> Loading files and directories
                    </div>
                ) : isErrorLike(treeOrError) ? (
                    // If the tree is actually a blob, be helpful and redirect to the blob page.
                    isNotTreeError(treeOrError) ? (
                        <Redirect to={toPrettyBlobURL({ repoName: repo.name, revision, commitID, filePath })} />
                    ) : (
                        <ErrorAlert error={treeOrError} />
                    )
                ) : (
                    <>
                        <header>
                            {treeOrError.isRoot ? (
                                <>
                                    <PageHeader
                                        path={[{ icon: SourceRepositoryIcon, text: displayRepoName(repo.name) }]}
                                        className="mb-3 test-tree-page-title"
                                    />
                                    {repo.description && <p>{repo.description}</p>}
                                    <RepositoryRootLinks
                                        repo={repo}
                                        tree={treeOrError}
                                        revision={revision}
                                        codeIntelligenceEnabled={codeIntelligenceEnabled}
                                        batchChangesEnabled={batchChangesEnabled}
                                    />
                                </>
                            ) : (
                                <PageHeader
                                    path={[{ icon: FolderIcon, text: filePath }]}
                                    className="mb-3 test-tree-page-title"
                                />
                            )}
                        </header>

                        <ExtensionViewsSection
                            className={classNames('my-3', styles.section)}
                            telemetryService={props.telemetryService}
                            settingsCascade={settingsCascade}
                            platformContext={props.platformContext}
                            extensionsController={props.extensionsController}
                            where="directory"
                            uri={uri}
                        />

                        <section className={classNames('test-tree-entries mb-3', styles.section)}>
                            <h2>Files and directories</h2>
                            <TreeEntriesSection
                                parentPath={filePath}
                                entries={treeOrError.entries}
                                fileDecorationsByPath={fileDecorationsByPath}
                                isLightTheme={props.isLightTheme}
                            />
                        </section>

                        <ActionsContainer {...props} menu={ContributableMenu.DirectoryPage} empty={null}>
                            {items => (
                                <section className={styles.section}>
                                    <h2>Actions</h2>
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
                        </ActionsContainer>

                        <div className={styles.section}>
                            <h2>Changes</h2>
                            <TreeCommits repoID={repo.id} commitID={commitID} filePath={filePath} className="mt-2" />
                        </div>
                    </>
                )}
            </Container>
        </div>
    )
}

/**
 * Report whether the error is a GraphQL error indicating that the fetched tree entry is not a
 * directory (and is instead likely a blob).
 */
export function isNotTreeError(error: ErrorLike): boolean {
    // We don't have error names on GraphQL errors.
    return /not a directory/i.test(error.message)
}
