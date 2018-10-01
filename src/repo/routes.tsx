import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { ResizablePanel } from '../panel/Panel'
import { getModeFromPath } from '../util'
import { formatHash, isLegacyFragment, parseHash } from '../util/url'
import { CodeIntelStatusIndicator } from './actions/CodeIntelStatusIndicator'
import { BlobPage } from './blob/BlobPage'
import { RepositoryCommitsPage } from './commits/RepositoryCommitsPage'
import { FilePathBreadcrumb } from './FilePathBreadcrumb'
import { RepoHeaderContributionPortal } from './RepoHeaderContributionPortal'
import { RepoRevContainerContext, RepoRevContainerRoute } from './RepoRevContainer'
import { RepoRevSidebar } from './RepoRevSidebar'
import { TreePage } from './TreePage'

/** Dev feature flag to make benchmarking the file tree in isolation easier. */
const hideRepoRevContent = localStorage.getItem('hideRepoRevContent')

export const repoRevContainerRoutes: ReadonlyArray<RepoRevContainerRoute> = [
    ...['', '/-/:objectType(blob|tree)/:filePath+'].map(routePath => ({
        path: routePath,
        exact: routePath === '',
        render: (
            context: RepoRevContainerContext &
                RouteComponentProps<{
                    objectType: 'blob' | 'tree' | undefined
                    filePath: string | undefined
                }>
        ) => {
            const objectType: 'blob' | 'tree' = context.match.params.objectType || 'tree'
            const filePath = context.match.params.filePath || '' // empty string is root
            const mode = getModeFromPath(filePath)

            // For blob pages with legacy URL fragment hashes like "#L17:19-21:23$foo:bar"
            // redirect to the modern URL fragment hashes like "#L17:19-21:23&tab=foo:bar"
            if (!hideRepoRevContent && objectType === 'blob' && isLegacyFragment(window.location.hash)) {
                const hash = parseHash(window.location.hash)
                const newHash = new URLSearchParams()
                if (hash.viewState) {
                    newHash.set('tab', hash.viewState)
                }
                return <Redirect to={window.location.pathname + window.location.search + formatHash(hash, newHash)} />
            }

            return (
                <>
                    {filePath && (
                        <>
                            <RepoHeaderContributionPortal
                                position="nav"
                                priority={10}
                                element={
                                    <FilePathBreadcrumb
                                        key="path"
                                        repoPath={context.repo.name}
                                        rev={context.rev}
                                        filePath={filePath}
                                        isDir={objectType === 'tree'}
                                    />
                                }
                                repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
                            />
                            {objectType === 'blob' && (
                                <RepoHeaderContributionPortal
                                    position="right"
                                    priority={-10}
                                    element={
                                        <CodeIntelStatusIndicator
                                            key="code-intel-status"
                                            userIsSiteAdmin={!!context.user && context.user.siteAdmin}
                                            repoPath={context.repo.name}
                                            rev={context.rev}
                                            commitID={context.resolvedRev.commitID}
                                            filePath={filePath}
                                            mode={mode}
                                        />
                                    }
                                    repoHeaderContributionsLifecycleProps={
                                        context.repoHeaderContributionsLifecycleProps
                                    }
                                />
                            )}
                        </>
                    )}
                    <RepoRevSidebar
                        className="repo-rev-container__sidebar"
                        repoID={context.repo.id}
                        repoPath={context.repo.name}
                        rev={context.rev}
                        commitID={context.resolvedRev.commitID}
                        filePath={context.match.params.filePath || '' || ''}
                        isDir={objectType === 'tree'}
                        defaultBranch={context.resolvedRev.defaultBranch || 'HEAD'}
                        history={context.history}
                        location={context.location}
                        extensionsController={context.extensionsController}
                    />
                    {!hideRepoRevContent && (
                        <div className="repo-rev-container__content">
                            {objectType === 'blob' ? (
                                <BlobPage
                                    repoPath={context.repo.name}
                                    repoID={context.repo.id}
                                    commitID={context.resolvedRev.commitID}
                                    rev={context.rev}
                                    filePath={context.match.params.filePath || ''}
                                    mode={mode}
                                    repoHeaderContributionsLifecycleProps={
                                        context.repoHeaderContributionsLifecycleProps
                                    }
                                    configurationCascade={context.configurationCascade}
                                    extensions={context.extensions}
                                    extensionsOnVisibleTextDocumentsChange={
                                        context.extensionsOnVisibleTextDocumentsChange
                                    }
                                    extensionsController={context.extensionsController}
                                    location={context.location}
                                    history={context.history}
                                    isLightTheme={context.isLightTheme}
                                    user={context.user}
                                />
                            ) : (
                                <TreePage
                                    repoPath={context.repo.name}
                                    repoID={context.repo.id}
                                    repoDescription={context.repo.description}
                                    commitID={context.resolvedRev.commitID}
                                    rev={context.rev}
                                    filePath={context.match.params.filePath || ''}
                                    configurationCascade={context.configurationCascade}
                                    extensionsController={context.extensionsController}
                                    extensions={context.extensions}
                                    location={context.location}
                                    history={context.history}
                                    isLightTheme={context.isLightTheme}
                                    onHelpPopoverToggle={context.onHelpPopoverToggle}
                                />
                            )}
                            <ResizablePanel {...context} />
                        </div>
                    )}
                </>
            )
        },
    })),
    {
        path: '/-/commits',
        render: context => (
            <RepositoryCommitsPage
                {...context}
                commitID={context.resolvedRev.commitID}
                repoHeaderContributionsLifecycleProps={context.repoHeaderContributionsLifecycleProps}
            />
        ),
    },
]
