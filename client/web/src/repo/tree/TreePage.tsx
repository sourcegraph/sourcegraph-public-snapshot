import React, { useMemo, useEffect, useState } from 'react'

import { mdiCog, mdiFolder, mdiSourceRepository } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { Redirect, Route, Switch, useRouteMatch } from 'react-router-dom'
import { catchError } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, encodeURIPathComponent, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { SearchContextProps } from '@sourcegraph/search'
import { fetchTreeEntries } from '@sourcegraph/shared/src/backend/repo'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
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
    Text,
} from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { BatchChangesIcon } from '../../batches/icons'
import { CodeIntelligenceProps } from '../../codeintel'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { ActionItemsBarProps } from '../../extensions/components/ActionItemsBar'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { RepositoryFields } from '../../graphql-operations'
import { basename } from '../../util/path'
import { RepositoryCompareArea } from '../compare/RepositoryCompareArea'
import { RepoRevisionWrapper } from '../components/RepoRevision'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { RepositoryFileTreePageProps } from '../RepositoryFileTreePage'
import { RepositoryGitDataContainer } from '../RepositoryGitDataContainer'
import { RepoCommits } from '../routes'
import { RepositoryStatsContributorsPage } from '../stats/RepositoryStatsContributorsPage'

import { RepositoryBranchesTab } from './BranchesTab'
import { HomeTab } from './HomeTab'
import { RepositoryTagTab } from './TagTab'
import { TreeNavigation } from './TreeNavigation'
import { TreePageContent } from './TreePageContent'
import { TreeTabList } from './TreeTabList'

import styles from './TreePage.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        TelemetryProps,
        ActivationProps,
        CodeIntelligenceProps,
        BatchChangesProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        BreadcrumbSetters {
    repo: RepositoryFields | undefined
    repoName: string
    /** The tree's path in TreePage. We call it filePath for consistency elsewhere. */
    filePath: string
    commitID: string
    revision: string
    location: H.Location
    history: H.History
    globbing: boolean
    useActionItemsBar: ActionItemsBarProps['useActionItemsBar']
    match: RepositoryFileTreePageProps['match']
    isSourcegraphDotCom: boolean
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

export const TreePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    repoName,
    commitID,
    revision,
    filePath,
    settingsCascade,
    useBreadcrumb,
    codeIntelligenceEnabled,
    batchChangesEnabled,
    useActionItemsBar,
    match,
    isSourcegraphDotCom,
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
                        telemetryService={props.telemetryService}
                    />
                ),
            }
        }, [filePath, repoName, revision, props.telemetryService])
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
                    requestGraphQL: props.platformContext.requestGraphQL,
                }).pipe(catchError((error): [ErrorLike] => [asError(error)])),
            [repoName, commitID, revision, filePath, props.platformContext]
        )
    )

    const showCodeInsights =
        !isErrorLike(settingsCascade.final) &&
        !!settingsCascade.final?.experimentalFeatures?.codeInsights &&
        settingsCascade.final['insights.displayLocation.directory'] === true

    // Add DirectoryViewer
    const uri = toURIWithPath({ repoName, commitID, filePath })

    const { extensionsController } = props
    useEffect(() => {
        if (!showCodeInsights || extensionsController === null) {
            return
        }

        const viewerIdPromise = extensionsController.extHostAPI
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
            Promise.all([extensionsController.extHostAPI, viewerIdPromise])
                .then(([extensionHostAPI, viewerId]) => {
                    if (viewerId) {
                        return extensionHostAPI.removeViewer(viewerId)
                    }
                    return
                })
                .catch(error => console.error('Error removing viewer from extension host:', error))
        }
    }, [uri, showCodeInsights, extensionsController])

    const getPageTitle = (): string => {
        const repoString = displayRepoName(repoName)
        if (filePath) {
            return `${basename(filePath)} - ${repoString}`
        }
        return `${repoString}`
    }

    // To start using the feature flag bellow, you can go to /site-admin/feature-flags and
    // create a new featureFlag named 'new-repo-page' and set its value to true.
    // https://docs.sourcegraph.com/dev/how-to/use_feature_flags#create-a-feature-flag
    const [isNewRepoPageEnabled] = useFeatureFlag('new-repo-page')

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
    const [showPageTitle, setShowPageTitle] = useState(true)
    const { path } = useRouteMatch()

    useMemo(() => {
        if (isNewRepoPageEnabled && treeOrError && !isErrorLike(treeOrError)) {
            setShowPageTitle(false)

            switch (path) {
                case `${treeOrError.url}/-/tag/tab`:
                    setSelectedTab('tags')
                    break
                case `${treeOrError.url}/-/docs/tab/:pathID*`:
                    setSelectedTab('docs')
                    setShowPageTitle(true)
                    break
                case `${treeOrError.url}/-/commits/tab`:
                    setSelectedTab('commits')
                    break
                case `${treeOrError.url}/-/branch/tab`:
                    setSelectedTab('branch')
                    break
                case `${treeOrError.url}/-/contributors/tab`:
                    setSelectedTab('contributors')
                    break
                case `${treeOrError.url}/-/compare/tab/:spec*`:
                    setSelectedTab('compare')
                    break
                case `${treeOrError.url}`:
                    setSelectedTab('home')
                    setShowPageTitle(true)
                    break
            }
        }
    }, [isNewRepoPageEnabled, path, treeOrError])

    const RootHeaderSection = ({ tree }: { tree: TreeFields }): React.ReactElement => (
        <>
            <div className="d-flex justify-content-between align-items-center">
                <div>
                    <PageHeader
                        path={[{ icon: mdiSourceRepository, text: displayRepoName(repoName) }]}
                        className="mb-3 test-tree-page-title"
                    />
                    {repo?.description && <Text>{repo.description}</Text>}
                </div>
                {isNewRepoPageEnabled && (
                    <ButtonGroup>
                        {!isSourcegraphDotCom && batchChangesEnabled && (
                            <Button
                                to="/batch-changes/create"
                                variant="secondary"
                                outline={true}
                                as={Link}
                                className="ml-1"
                            >
                                <Icon as={BatchChangesIcon} aria-hidden={true} /> Create batch change
                            </Button>
                        )}

                        {repo?.viewerCanAdminister && (
                            <Button
                                to={`/${encodeURIPathComponent(repoName)}/-/settings`}
                                variant="secondary"
                                outline={true}
                                as={Link}
                                className="ml-1"
                                aria-label="Repository settings"
                            >
                                <Icon aria-hidden={true} svgPath={mdiCog} />
                            </Button>
                        )}
                    </ButtonGroup>
                )}
            </div>
            {isNewRepoPageEnabled ? (
                <TreeTabList tree={tree} selectedTab={selectedTab} setSelectedTab={setSelectedTab} />
            ) : (
                <TreeNavigation
                    batchChangesEnabled={batchChangesEnabled}
                    codeIntelligenceEnabled={codeIntelligenceEnabled}
                    repoName={repoName}
                    viewerCanAdminister={repo?.viewerCanAdminister}
                    revision={revision}
                    tree={tree}
                />
            )}
        </>
    )

    return (
        <div className={styles.treePage}>
            <Container className={styles.container}>
                {!showPageTitle && <PageTitle title={getPageTitle()} />}
                {treeOrError === undefined || repo === undefined ? (
                    <div>
                        <LoadingSpinner /> Loading files and directories
                    </div>
                ) : isErrorLike(treeOrError) ? (
                    // If the tree is actually a blob, be helpful and redirect to the blob page.
                    // We don't have error names on GraphQL errors.
                    /not a directory/i.test(treeOrError.message) ? (
                        <Redirect to={toPrettyBlobURL({ repoName, revision, commitID, filePath })} />
                    ) : (
                        <ErrorAlert error={treeOrError} />
                    )
                ) : (
                    <div className={classNames(styles.header)}>
                        <header className="mb-3">
                            {treeOrError.isRoot ? (
                                <RootHeaderSection tree={treeOrError} />
                            ) : (
                                <PageHeader
                                    path={[{ icon: mdiFolder, text: filePath }]}
                                    className="mb-3 mr-2 test-tree-page-title"
                                />
                            )}
                        </header>

                        {isNewRepoPageEnabled ? (
                            <div>
                                <section className={classNames('test-tree-entries mb-3', styles.section)}>
                                    <Switch>
                                        <Route
                                            path={`${treeOrError.url}/-/tag/tab`}
                                            render={routeComponentProps => (
                                                <RepositoryTagTab repo={repo} {...routeComponentProps} />
                                            )}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/commits/tab`}
                                            render={routeComponentProps => (
                                                <RepoCommits
                                                    repo={repo}
                                                    revision={revision || ''}
                                                    useBreadcrumb={useBreadcrumb}
                                                    {...props}
                                                    {...routeComponentProps}
                                                />
                                            )}
                                        />
                                        <Route
                                            path={`${treeOrError.url}`}
                                            exact={true}
                                            render={routeComponentProps => (
                                                <HomeTab
                                                    {...homeTabProps}
                                                    {...props}
                                                    {...routeComponentProps}
                                                    repo={repo}
                                                />
                                            )}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/branch/tab`}
                                            render={routeComponentProps => (
                                                <RepositoryBranchesTab repo={repo} {...routeComponentProps} />
                                            )}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/contributors/tab`}
                                            render={routeComponentProps => (
                                                <RepositoryStatsContributorsPage
                                                    {...routeComponentProps}
                                                    repo={repo}
                                                    {...props}
                                                />
                                            )}
                                        />
                                        <Route
                                            path={`${treeOrError.url}/-/compare/tab`}
                                            render={() => (
                                                <RepoRevisionWrapper>
                                                    <RepositoryGitDataContainer
                                                        {...props}
                                                        repoName={repoName}
                                                        resolvedRevisionOrError={undefined}
                                                    >
                                                        <RepositoryCompareArea
                                                            repo={repo}
                                                            match={match}
                                                            settingsCascade={settingsCascade}
                                                            useBreadcrumb={useBreadcrumb}
                                                            {...props}
                                                        />
                                                    </RepositoryGitDataContainer>
                                                </RepoRevisionWrapper>
                                            )}
                                        />
                                    </Switch>
                                </section>
                            </div>
                        ) : (
                            <TreePageContent
                                filePath={filePath}
                                tree={treeOrError}
                                repo={repo}
                                revision={revision}
                                commitID={commitID}
                                {...props}
                            />
                        )}
                    </div>
                )}
            </Container>
        </div>
    )
}
