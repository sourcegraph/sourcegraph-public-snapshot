import classNames from 'classnames'
import React, { useEffect, useMemo } from 'react'
import { Redirect } from 'react-router'

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
    SourceSetAtTreePageVariables,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    SourceSetAtTreePageResult,
    SourceSetAtTreeFields,
    PrimaryComponentForTreeFields,
} from '../../../../graphql-operations'
import { ComponentActionPopoverButton } from '../../../../repo/actions/source-set-view-mode-action/SourceSetViewModeAction'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { isNotTreeError, TreePage, useTreePageBreadcrumb } from '../../../../repo/tree/TreePage'
import treePageStyles from '../../../../repo/tree/TreePage.module.scss'
import { basename } from '../../../../util/path'
import { CatalogPage, CatalogPage2 } from '../../components/catalog-area-header/CatalogPage'
import { COMPONENT_TAG_FRAGMENT } from '../../components/component-tag/ComponentTag'

import styles from './SourceSetAtTreePage.module.scss'
import { SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT } from './SourceSetDescendentComponents'
import { CodeTab } from './tabs/code/CodeTab'
import { SOURCE_SET_README_FRAGMENT } from './tabs/code/ComponentReadme'
import { COMPONENT_OWNER_FRAGMENT } from './tabs/code/sidebar/ComponentOwnerSidebarItem'
import { SOURCE_SET_CODE_OWNERS_FRAGMENT } from './tabs/code/sidebar/SourceSetCodeOwnersSidebarItem'
import { SOURCE_SET_CONTRIBUTORS_FRAGMENT } from './tabs/code/sidebar/SourceSetContributorsSidebarItem'
import { SOURCE_SET_FILES_FRAGMENT } from './tabs/code/SourceSetTreeEntries'
import { CatalogRelations } from './tabs/graph/CatalogRelations'
import { UsageTab } from './tabs/usage/UsageTab'
import { WhoKnowsTab } from './tabs/who-knows/WhoKnowsTab'
import { useSourceSetAtTreeViewOptions } from './useSourceSetAtTreeViewOptions'

const TREE_OR_COMPONENT_PAGE = gql`
    query SourceSetAtTreePage($repo: ID!, $commitID: String!, $inputRevspec: String!, $path: String!) {
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
        ...SourceSetAtTreeFields
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
        ...SourceSetAtTreeFields
        ...ComponentOwnerFields
        tags {
            ...ComponentTagFields
        }
    }

    fragment SourceSetAtTreeFields on SourceSet {
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

/**
 * A page that fetches and displays the SourceSet at a given tree. If there is a component whose
 * primary location is the tree, then the component is displayed. Otherwise, a trivial single-tree
 * SourceSet is displayed.
 *
 * If the user wants to view the tree as "just a tree" and not see additional information from the
 * component that lives at the tree, the user can select to "View as tree".
 */
export const SourceSetAtTreePage: React.FunctionComponent<React.ComponentPropsWithoutRef<typeof TreePage>> = ({
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

    const { data, error, loading } = useQuery<SourceSetAtTreePageResult, SourceSetAtTreePageVariables>(
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
                <SourceSetAtTree
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
    data: Extract<SourceSetAtTreePageResult['node'], { __typename: 'Repository' }>
}

const tabContentClassName = classNames('flex-1 align-self-stretch', styles.tabContent)

const SourceSetAtTree: React.FunctionComponent<Props> = ({
    filePath,
    repoID,
    repository,
    tree,
    primaryComponent,
    data,
    useBreadcrumb,
    ...props
}) => {
    const sourceSetAtTreeViewOptions = useSourceSetAtTreeViewOptions()

    const sourceSet: SourceSetAtTreeFields =
        sourceSetAtTreeViewOptions.sourceSetAtTreeViewMode === 'auto' ? primaryComponent ?? tree : tree

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
                                  {...sourceSetAtTreeViewOptions}
                              />
                          ),
                          divider: <span className="mx-1" />,
                      }
                    : null,
            [primaryComponent, sourceSetAtTreeViewOptions]
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
                            {...sourceSetAtTreeViewOptions}
                            repository={repository}
                            tree={tree}
                            component={primaryComponent}
                            sourceSet={sourceSet}
                            useHash={true}
                            className={classNames('py-3', tabContentClassName)}
                        />
                    ),
                },
                {
                    path: 'who-knows',
                    text: 'Who knows?',
                    content: (
                        <WhoKnowsTab
                            {...props}
                            sourceSet={sourceSet.id}
                            className={classNames('py-3', tabContentClassName)}
                        />
                    ),
                },
                sourceSet && sourceSet.__typename === 'Component'
                    ? {
                          path: 'graph',
                          text: 'Graph',
                          content: (
                              <div className={classNames('py-3', tabContentClassName)}>
                                  <CatalogRelations component={sourceSet.id} useURLForConnectionParams={true} />
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
        [primaryComponent, props, repository, sourceSet, tree, sourceSetAtTreeViewOptions]
    )

    return <CatalogPage2 tabs={tabs} useHash={true} tabsClassName={styles.tabs} />
}
