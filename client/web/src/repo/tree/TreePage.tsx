import React, { useMemo, useEffect, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import CodeJsonIcon from 'mdi-react/CodeJsonIcon'
import FolderIcon from 'mdi-react/FolderIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { Redirect, Route, Switch } from 'react-router-dom'
import { EMPTY } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, encodeURIPathComponent, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { SearchContextProps } from '@sourcegraph/search'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { toURIWithPath, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    useObservable,
    Link,
    Icon,
    ButtonGroup,
    Button,
} from '@sourcegraph/wildcard'

import { getFileDecorations } from '../../backend/features'
import { BatchChangesProps } from '../../batches'
import { BatchChangesIcon } from '../../batches/icons'
import { CodeIntelligenceProps } from '../../codeintel'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { ActionItemsBar, ActionItemsBarProps } from '../../extensions/components/ActionItemsBar'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { RepositoryFields } from '../../graphql-operations'
import { basename } from '../../util/path'
import { fetchTreeEntries } from '../backend'
import { RepositoryCompareArea } from '../compare/RepositoryCompareArea'
import { RepoRevisionWrapper } from '../components/RepoRevision'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { RepositoryFileTreePageProps } from '../RepositoryFileTreePage'
import { RepositoryGitDataContainer } from '../RepositoryGitDataContainer'
import { RepoCommits, RepoDocs } from '../routes'
import { RepositoryStatsContributorsPage } from '../stats/RepositoryStatsContributorsPage'

import { RepositoryBranchesTab } from './BranchesTab'
import { HomeTab } from './HomeTab'
import { RepositoryTagTab } from './TagTab'
import { TreeEntriesSection } from './TreeEntriesSection'
import { TreeNavigation } from './TreeNavigation'
import { TreeTabList } from './TreeTabList'

import styles from './TreePage.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        FeatureFlagProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        CodeIntelligenceProps,
        BatchChangesProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        BreadcrumbSetters {
    repo: RepositoryFields
    /** The tree's path in TreePage. We call it filePath for consistency elsewhere. */
    filePath: string
    commitID: string
    revision: string
    location: H.Location
    history: H.History
    globbing: boolean
    useActionItemsBar: ActionItemsBarProps['useActionItemsBar']
    match: RepositoryFileTreePageProps['match']
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
    settingsCascade,
    useBreadcrumb,
    codeIntelligenceEnabled,
    batchChangesEnabled,
    useActionItemsBar,
    match,
    featureFlags,
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

    // To start using the feature flag bellow, you can go to /site-admin/feature-flags and
    // create a new featurFlag named 'new-repo-page' and set its value to true.
    // https://docs.sourcegraph.com/dev/how-to/use_feature_flags#create-a-feature-flag

    const newRepoPage = featureFlags.get('new-repo-page')

    const homeTabProps = {
        repo,
        commitID,
        revision,
        filePath,
        settingsCascade,
        codeIntelligenceEnabled,
        batchChangesEnabled,
        location,
    }

    const [selectedTab, setSelectedTab] = useState('home')

    return (
        <div className={styles.treePage}>
            <Container className={styles.container}>
                <PageTitle title={getPageTitle()} />
                {treeOrError === undefined ? (
                    <div>
                        <LoadingSpinner /> Loading files and directories
                    </div>
                ) : isErrorLike(treeOrError) ? (
                    // If the tree is actually a blob, be helpful and redirect to the blob page.
                    // We don't have error names on GraphQL errors.
                    /not a directory/i.test(treeOrError.message) ? (
                        <Redirect to={toPrettyBlobURL({ repoName: repo.name, revision, commitID, filePath })} />
                    ) : (
                        <ErrorAlert error={treeOrError} />
                    )
                ) : (
                    <div className={classNames(styles.header)}>
                        <header className="mb-3">
                            {treeOrError.isRoot ? (
                                <>
                                    <div className="d-flex justify-content-between align-items-center">
                                        <div>
                                            <PageHeader
                                                path={[
                                                    { icon: SourceRepositoryIcon, text: displayRepoName(repo.name) },
                                                ]}
                                                className="mb-3 test-tree-page-title"
                                            />
                                            {repo.description && <p>{repo.description}</p>}
                                        </div>
                                        <ButtonGroup>
                                            <Button
                                                to={`/search?q=${encodeURIPathComponent(
                                                    `context:global count:all repo:dependencies(${repo.name.replaceAll(
                                                        '.',
                                                        '\\.'
                                                    )}$) `
                                                )}`}
                                                variant="secondary"
                                                as={Link}
                                                className="ml-1"
                                            >
                                                <Icon as={CodeJsonIcon} /> Search Dependencies
                                            </Button>

                                            {!batchChangesEnabled &&(
                                                <Button
                                                    to="/batch-changes"
                                                    variant="secondary"
                                                    as={Link}
                                                    className="ml-1"
                                                >
                                                    <Icon as={BatchChangesIcon} /> Create batch change
                                                </Button>
                                            )}

                                            {repo.viewerCanAdminister && (
                                                <Button
                                                    to={`/${encodeURIPathComponent(repo.name)}/-/settings`}
                                                    variant="secondary"
                                                    as={Link}
                                                    className="ml-1"
                                                >
                                                    <Icon as={SettingsIcon} />
                                                </Button>
                                            )}
                                        </ButtonGroup>
                                    </div>
                                    {!newRepoPage && (
                                        <TreeNavigation
                                            batchChangesEnabled={batchChangesEnabled}
                                            codeIntelligenceEnabled={codeIntelligenceEnabled}
                                            repo={repo}
                                            revision={revision}
                                            tree={treeOrError}
                                        />
                                    )}
                                    <TreeTabList
                                        tree={treeOrError}
                                        selectedTab={selectedTab}
                                        setSelectedTab={setSelectedTab}
                                    />
                                </>
                            ) : (
                                <PageHeader
                                    path={[{ icon: FolderIcon, text: filePath }]}
                                    className="mb-3 test-tree-page-title"
                                />
                            )}
                        </header>

                        <div>
                            <section className={classNames('test-tree-entries mb-3', styles.section)}>
                                {treeOrError.isRoot && (
                                    <Switch>
                                        <Route
                                            path={`${treeOrError.url}/-/tag/tab`}
                                            render={routeComponentProps => {
                                                setSelectedTab('tags')
                                                return <RepositoryTagTab repo={repo} {...routeComponentProps} />
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/docs/tab`}
                                            render={routeComponentProps => {
                                                setSelectedTab('docs')
                                                return (
                                                    <RepoDocs
                                                        repo={repo}
                                                        useBreadcrumb={useBreadcrumb}
                                                        {...routeComponentProps}
                                                        {...props}
                                                    />
                                                )
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/commits/tab`}
                                            render={routeComponentProps => {
                                                setSelectedTab('commits')
                                                return (
                                                    <RepoCommits
                                                        repo={repo}
                                                        useBreadcrumb={useBreadcrumb}
                                                        {...props}
                                                        {...routeComponentProps}
                                                    />
                                                )
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}`}
                                            exact={true}
                                            render={routeComponentProps => {
                                                setSelectedTab('home')
                                                return (
                                                    <HomeTab
                                                        {...homeTabProps}
                                                        {...props}
                                                        {...routeComponentProps}
                                                        repo={repo}
                                                    />
                                                )
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/branch/tab`}
                                            render={routeComponentProps => {
                                                setSelectedTab('branch')
                                                return <RepositoryBranchesTab repo={repo} {...routeComponentProps} />
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/contributors/tab`}
                                            render={routeComponentProps => {
                                                setSelectedTab('contributors')
                                                return (
                                                    <RepositoryStatsContributorsPage
                                                        {...routeComponentProps}
                                                        repo={repo}
                                                        {...props}
                                                    />
                                                )
                                            }}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/compare/tab`}
                                            render={() => {
                                                setSelectedTab('compare')
                                                return (
                                                    <RepoRevisionWrapper>
                                                        <RepositoryGitDataContainer {...props} repoName={repo.name}>
                                                            <RepositoryCompareArea
                                                                repo={repo}
                                                                match={match}
                                                                settingsCascade={settingsCascade}
                                                                useBreadcrumb={useBreadcrumb}
                                                                {...props}
                                                            />
                                                        </RepositoryGitDataContainer>
                                                        <ActionItemsBar
                                                            extensionsController={props.extensionsController}
                                                            platformContext={props.platformContext}
                                                            useActionItemsBar={useActionItemsBar}
                                                            location={props.location}
                                                            telemetryService={props.telemetryService}
                                                        />
                                                    </RepoRevisionWrapper>
                                                )
                                            }}
                                        />
                                    </Switch>
                                )}
                                {!treeOrError.isRoot && (
                                    <div className={styles.section}>
                                        <h2>Files and directories</h2>
                                        <TreeEntriesSection
                                            parentPath={filePath}
                                            entries={treeOrError.entries}
                                            fileDecorationsByPath={fileDecorationsByPath}
                                            isLightTheme={props.isLightTheme}
                                        />
                                    </div>
                                )}
                            </section>
                        </div>
                    </div>
                )}
            </Container>
        </div>
    )
}
