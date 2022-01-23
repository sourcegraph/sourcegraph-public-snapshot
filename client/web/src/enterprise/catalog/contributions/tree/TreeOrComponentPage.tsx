import classNames from 'classnames'
import { Location, LocationDescriptorObject } from 'history'
import React, { useEffect, useCallback, useMemo } from 'react'
import { Redirect, useHistory, useLocation } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { FileSpec, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { BreadcrumbSetters } from '@sourcegraph/web/src/components/Breadcrumbs'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import {
    TreeOrComponentPageVariables,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    TreeOrComponentPageResult,
    TreeOrComponentSourceSetFields,
    PrimaryComponentForTreeFields,
} from '../../../../graphql-operations'
import { ComponentActionPopoverButton } from '../../../../repo/actions/source-set-view-mode-action/SourceSetViewModeAction'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { isNotTreeError, TreePage, useTreePageBreadcrumb } from '../../../../repo/tree/TreePage'
import treePageStyles from '../../../../repo/tree/TreePage.module.scss'
import { basename } from '../../../../util/path'
import { CatalogPage, CatalogPage2 } from '../../components/catalog-area-header/CatalogPage'
import { COMPONENT_TAG_FRAGMENT } from '../../components/component-tag/ComponentTag'
import { CodeTab } from '../../pages/source-set/code/CodeTab'
import { SOURCE_SET_FILES_FRAGMENT } from '../../pages/source-set/code/SourceSetTreeEntries'
import { CatalogRelations } from '../../pages/source-set/graph/CatalogRelations'
import { COMPONENT_OWNER_FRAGMENT } from '../../pages/source-set/meta/ComponentOwnerSidebarItem'
import { SOURCE_SET_CODE_OWNERS_FRAGMENT } from '../../pages/source-set/meta/SourceSetCodeOwnersSidebarItem'
import { SOURCE_SET_CONTRIBUTORS_FRAGMENT } from '../../pages/source-set/meta/SourceSetContributorsSidebarItem'
import { SOURCE_SET_README_FRAGMENT } from '../../pages/source-set/readme/ComponentReadme'
import { UsageTab } from '../../pages/source-set/usage/UsageTab'
import { WhoKnowsTab } from '../../pages/source-set/who-knows/WhoKnowsTab'

import { SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT } from './SourceSetDescendentComponents'
import { TreeOrComponentHeader } from './TreeOrComponentHeader'
import styles from './TreeOrComponentPage.module.scss'

const TREE_OR_COMPONENT_PAGE = gql`
    query TreeOrComponentPage($repo: ID!, $commitID: String!, $inputRevspec: String!, $path: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                __typename
                id
                ...RepositoryForTreeFields
                commit(rev: $commitID, inputRevspec: $inputRevspec) {
                    id
                    tree(path: $path) {
                        ...TreeEntryForTreeFields
                    }
                }
                primaryComponents: components(path: $path, primary: true, recursive: false) {
                    ...PrimaryComponentForTreeFields
                }
            }
        }
    }

    fragment RepositoryForTreeFields on Repository {
        id
        name
        description
    }

    fragment TreeEntryForTreeFields on GitTree {
        path
        name
        isRoot
        url
        ...TreeOrComponentSourceSetFields
    }

    fragment PrimaryComponentForTreeFields on Component {
        __typename
        id
        name
        description
        kind
        lifecycle
        catalogURL
        url
        ...TreeOrComponentSourceSetFields
        ...ComponentOwnerFields
        tags {
            ...ComponentTagFields
        }
    }

    fragment TreeOrComponentSourceSetFields on SourceSet {
        id
        ...SourceSetDescendentComponentsFields
        ...SourceSetFilesFields
        ...SourceSetReadmeFields
        ...SourceSetCodeOwnersFields
        ...SourceSetContributorsFields
        branches(first: 0, interactive: false) {
            totalCount
        }
        commitsForLastCommit: commits(first: 1) {
            nodes {
                ...GitCommitFields
            }
        }
        usage {
            __typename
        }
    }

    ${SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT}
    ${SOURCE_SET_FILES_FRAGMENT}
    ${SOURCE_SET_README_FRAGMENT}
    ${SOURCE_SET_CODE_OWNERS_FRAGMENT}
    ${SOURCE_SET_CONTRIBUTORS_FRAGMENT}
    ${gitCommitFragment}
    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_TAG_FRAGMENT}
`

export const TreeOrComponentPage: React.FunctionComponent<React.ComponentPropsWithoutRef<typeof TreePage>> = ({
    repo,
    revision,
    commitID,
    filePath,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent(filePath === '' ? 'Repository' : 'Tree'), [
        filePath,
        props.telemetryService,
    ])
    useTreePageBreadcrumb({
        repo,
        revision,
        filePath,
        telemetryService: props.telemetryService,
        useBreadcrumb: props.useBreadcrumb,
    })

    const { data, error, loading } = useQuery<TreeOrComponentPageResult, TreeOrComponentPageVariables>(
        TREE_OR_COMPONENT_PAGE,
        {
            variables: { repo: repo.id, commitID, inputRevspec: revision, path: filePath },
            fetchPolicy: 'cache-first',
        }
    )

    const pageTitle = `${filePath ? `${basename(filePath)} - ` : ''}${displayRepoName(repo.name)}`

    if (error && isNotTreeError(error)) {
        return <Redirect to={toPrettyBlobURL({ repoName: repo.name, revision, commitID, filePath })} />
    }
    return (
        <Container className={classNames(treePageStyles.container, 'pt-0')}>
            <PageTitle title={pageTitle} />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <ErrorAlert error={error} />
            ) : !data || !data.node || data.node.__typename !== 'Repository' ? (
                <ErrorAlert error="Not a repository" />
            ) : !data.node.commit?.tree ? (
                <ErrorAlert error="404 Not Found" />
            ) : (
                <TreeOrComponent
                    {...props}
                    filePath={filePath}
                    repoID={repo.id}
                    repository={data.node}
                    tree={data.node.commit?.tree ?? null}
                    primaryComponent={data.node.primaryComponents[0] ?? null}
                    data={data.node}
                />
            )}
        </Container>
    )
}

interface Props
    extends SettingsCascadeProps,
        TelemetryProps,
        BreadcrumbSetters,
        ExtensionsControllerProps,
        ThemeProps,
        FileSpec {
    repoID: Scalars['ID']
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    primaryComponent: PrimaryComponentForTreeFields | null
    data: Extract<TreeOrComponentPageResult['node'], { __typename: 'Repository' }>
}

const tabContentClassName = classNames('flex-1 align-self-stretch', styles.tabContent)

const TreeOrComponent: React.FunctionComponent<Props> = ({
    filePath,
    repoID,
    repository,
    tree,
    primaryComponent,
    data,
    useBreadcrumb,
    ...props
}) => {
    const treeOrComponentViewOptions = useTreeOrComponentViewOptions()

    const sourceSet: TreeOrComponentSourceSetFields =
        treeOrComponentViewOptions.treeOrComponentViewMode === 'auto' ? primaryComponent ?? tree : tree

    useBreadcrumb(
        useMemo(
            () =>
                primaryComponent
                    ? {
                          key: 'component',
                          className: 'flex-shrink-past-contents align-self-stretch',
                          element: (
                              <ComponentActionPopoverButton
                                  component={primaryComponent}
                                  {...treeOrComponentViewOptions}
                              />
                          ),
                          divider: <span className="mx-1" />,
                      }
                    : null,
            [primaryComponent, treeOrComponentViewOptions]
        )
    )

    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: ['', 'contributors', 'code-owners', 'commits', 'branches'],
                    text: 'Code',
                    content: (
                        <CodeTab
                            {...props}
                            {...treeOrComponentViewOptions}
                            repository={repository}
                            tree={tree}
                            component={primaryComponent}
                            sourceSet={sourceSet}
                            useHash={true}
                            className={tabContentClassName}
                        />
                    ),
                },
                {
                    path: 'who-knows',
                    text: 'Who knows?',
                    content: <WhoKnowsTab {...props} sourceSet={sourceSet.id} className={tabContentClassName} />,
                },
                sourceSet && sourceSet.__typename === 'Component'
                    ? {
                          path: 'graph',
                          text: 'Graph',
                          content: (
                              <div className={classNames('p-3', tabContentClassName)}>
                                  <CatalogRelations
                                      component={sourceSet.id}
                                      useURLForConnectionParams={true}
                                      className="mb-3"
                                  />
                              </div>
                          ),
                      }
                    : null,
                sourceSet?.usage && {
                    path: 'usage',
                    text: 'Usage',
                    content: <UsageTab {...props} sourceSet={sourceSet.id} className={tabContentClassName} />,
                },
            ].filter(isDefined),
        [primaryComponent, props, repository, sourceSet, tree, treeOrComponentViewOptions]
    )

    return (
        <CatalogPage2
            header={
                <TreeOrComponentHeader
                    repository={repository}
                    tree={tree}
                    primaryComponent={primaryComponent}
                    {...treeOrComponentViewOptions}
                />
            }
            tabs={tabs}
            useHash={true}
            tabsClassName={styles.tabs}
        />
    )
}

type TreeOrComponentViewMode = 'auto' | 'tree'

export interface TreeOrComponentViewOptionsProps {
    treeOrComponentViewMode: TreeOrComponentViewMode
    treeOrComponentViewModeURL: Record<TreeOrComponentViewMode, LocationDescriptorObject>
    setTreeOrComponentViewMode: (mode: TreeOrComponentViewMode) => void
}

export function useTreeOrComponentViewOptions(): TreeOrComponentViewOptionsProps {
    const location = useLocation()
    const history = useHistory()

    const treeOrComponentViewMode: TreeOrComponentViewMode = useMemo(
        () => (new URLSearchParams(location.search).get('as') === 'tree' ? 'tree' : 'auto'),
        [location.search]
    )
    const treeOrComponentViewModeURL = useMemo<TreeOrComponentViewOptionsProps['treeOrComponentViewModeURL']>(
        () => ({
            auto: makeTreeOrComponentViewURL(location, 'auto'),
            tree: makeTreeOrComponentViewURL(location, 'tree'),
        }),
        [location]
    )
    const setTreeOrComponentViewMode = useCallback<TreeOrComponentViewOptionsProps['setTreeOrComponentViewMode']>(
        mode => history.push(makeTreeOrComponentViewURL(location, mode)),
        [history, location]
    )

    return useMemo(() => ({ treeOrComponentViewMode, treeOrComponentViewModeURL, setTreeOrComponentViewMode }), [
        setTreeOrComponentViewMode,
        treeOrComponentViewMode,
        treeOrComponentViewModeURL,
    ])
}

function makeTreeOrComponentViewURL(location: Location, mode: TreeOrComponentViewMode): LocationDescriptorObject {
    const search = new URLSearchParams(location.search)
    if (mode === 'tree') {
        search.set('as', mode)
    } else {
        search.delete('as')
    }

    return { ...location, search: search.toString() }
}
