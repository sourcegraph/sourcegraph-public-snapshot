import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import formatISO from 'date-fns/formatISO'
import subYears from 'date-fns/subYears'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ContributableMenu } from '@sourcegraph/client-api'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Heading, Link, useObservable } from '@sourcegraph/wildcard'

import { getFileDecorations } from '../../backend/features'
import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import {
    GitCommitFields,
    TreeCommitsResult,
    TreeCommitsVariables,
    TreePageRepositoryFields,
} from '../../graphql-operations'
import { GitCommitNode } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'

import { TreeEntriesSection } from './TreeEntriesSection'

import styles from './TreePage.module.scss'

const TREE_COMMITS_PER_PAGE = 10

const TREE_COMMITS_QUERY = gql`
    query TreeCommits(
        $repo: ID!
        $revspec: String!
        $first: Int
        $filePath: String
        $after: String
        $afterCursor: String
    ) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    ancestors(first: $first, path: $filePath, after: $after, afterCursor: $afterCursor) {
                        nodes {
                            ...GitCommitFields
                        }
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                    }
                }
            }
        }
    }
    ${gitCommitFragment}
`

interface TreePageContentProps extends ExtensionsControllerProps, ThemeProps, TelemetryProps, PlatformContextProps {
    filePath: string
    tree: TreeFields
    repo: TreePageRepositoryFields
    commitID: string
    location: H.Location
    revision: string
}

export const TreePageContent: React.FunctionComponent<React.PropsWithChildren<TreePageContentProps>> = ({
    filePath,
    tree,
    repo,
    commitID,
    revision,
    ...props
}) => {
    const [showOlderCommits, setShowOlderCommits] = useState(false)
    const after = useMemo(() => (showOlderCommits ? null : formatISO(subYears(Date.now(), 1))), [showOlderCommits])

    const { connection, error, loading, hasNextPage, fetchMore, refetchAll } = useShowMorePagination<
        TreeCommitsResult,
        TreeCommitsVariables,
        GitCommitFields
    >({
        query: TREE_COMMITS_QUERY,
        variables: {
            repo: repo.id,
            revspec: revision || '',
            first: TREE_COMMITS_PER_PAGE,
            filePath,
            after,
            afterCursor: null,
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                return { nodes: [] }
            }
            if (node.__typename !== 'Repository') {
                return { nodes: [] }
            }
            if (!node.commit?.ancestors) {
                return { nodes: [] }
            }

            return node.commit.ancestors
        },
        options: {
            fetchPolicy: 'cache-first',
            useAlternateAfterCursor: true,
        },
    })

    // We store the refetchAll callback in a ref since it will update when
    // variables or result length change and we need to call an up-to-date
    // version in the useEffect below to refetch the proper results.
    //
    // TODO: See if we can make refetchAll stable
    const refetchAllRef = useRef(refetchAll)
    useEffect(() => {
        refetchAllRef.current = refetchAll
    }, [refetchAll])

    useEffect(() => {
        if (showOlderCommits && refetchAllRef.current) {
            // Updating the variables alone is not enough to force a loading
            // indicator to show, so we need to refetch the results.
            refetchAllRef.current()
        }
    }, [showOlderCommits])

    const fileDecorationsByPath =
        useObservable<FileDecorationsByPath>(
            useMemo(
                () =>
                    getFileDecorations({
                        files: tree.entries,
                        extensionsController: props.extensionsController,
                        repoName: repo.name,
                        commitID,
                        parentNodeUri: tree.url,
                    }),
                [commitID, props.extensionsController, repo.name, tree.entries, tree.url]
            )
        ) ?? {}

    const onShowOlderCommitsClicked = useCallback((event: React.MouseEvent): void => {
        event.preventDefault()
        setShowOlderCommits(true)
    }, [])

    const showAllCommits = (
        <Button
            className="test-tree-page-show-all-commits"
            onClick={onShowOlderCommitsClicked}
            variant="secondary"
            size="sm"
        >
            Show commits older than one year
        </Button>
    )

    const { extensionsController } = props

    const showLinkToCommitsPage = connection && hasNextPage && connection.nodes.length > TREE_COMMITS_PER_PAGE

    return (
        <>
            <section className={classNames('test-tree-entries mb-3', styles.section)}>
                <Heading as="h3" styleAs="h2">
                    Files and directories
                </Heading>
                <TreeEntriesSection
                    parentPath={filePath}
                    entries={tree.entries}
                    fileDecorationsByPath={fileDecorationsByPath}
                    isLightTheme={props.isLightTheme}
                />
            </section>
            {extensionsController !== null && window.context.enableLegacyExtensions ? (
                <ActionsContainer
                    {...props}
                    extensionsController={extensionsController}
                    menu={ContributableMenu.DirectoryPage}
                    empty={null}
                >
                    {items => (
                        <section className={styles.section}>
                            <Heading as="h3" styleAs="h2">
                                Actions
                            </Heading>
                            {items.map(item => (
                                <Button
                                    {...props}
                                    extensionsController={extensionsController}
                                    key={item.action.id}
                                    {...item}
                                    className="mr-1 mb-1"
                                    variant="secondary"
                                    as={ActionItem}
                                />
                            ))}
                        </section>
                    )}
                </ActionsContainer>
            ) : null}

            <ConnectionContainer className={styles.section}>
                <Heading as="h3" styleAs="h2">
                    Changes
                </Heading>

                {error && <ErrorAlert error={error} className="w-100 mb-0" />}
                <ConnectionList className="list-group list-group-flush w-100">
                    {connection?.nodes.map(node => (
                        <GitCommitNode
                            key={node.id}
                            className={classNames('list-group-item', styles.gitCommitNode)}
                            messageSubjectClassName={styles.gitCommitNodeMessageSubject}
                            compact={true}
                            wrapperElement="li"
                            node={node}
                        />
                    ))}
                </ConnectionList>
                {loading && <ConnectionLoading />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            centered={true}
                            first={TREE_COMMITS_PER_PAGE}
                            connection={connection}
                            noun={showOlderCommits ? 'commit' : 'commit in the past year'}
                            pluralNoun={showOlderCommits ? 'commits' : 'commits in the past year'}
                            hasNextPage={hasNextPage}
                            emptyElement={null}
                        />
                        {hasNextPage ? (
                            showLinkToCommitsPage ? (
                                <Link to={`${repo.url}/-/commits${filePath ? `/${filePath}` : ''}`}>
                                    Show all commits
                                </Link>
                            ) : (
                                <ShowMoreButton centered={true} onClick={fetchMore} />
                            )
                        ) : null}
                        {!hasNextPage && !showOlderCommits ? showAllCommits : null}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}
