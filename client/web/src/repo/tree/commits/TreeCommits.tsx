import React, { useCallback, useState, useEffect, useMemo, useRef } from 'react'

import classNames from 'classnames'
import formatISO from 'date-fns/formatISO'
import startOfDay from 'date-fns/startOfDay'
import subYears from 'date-fns/subYears'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ErrorAlert, Button, Heading, Link } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    SummaryContainer,
    ConnectionList,
    ConnectionLoading,
    ShowMoreButton,
    ConnectionSummary,
} from '../../../components/FilteredConnection/ui'
import {
    GitCommitFields,
    TreeCommitsResult,
    TreeCommitsVariables,
    TreePageRepositoryFields,
} from '../../../graphql-operations'
import { GitCommitNode } from '../../commits/GitCommitNode'
import { gitCommitFragment } from '../../commits/RepositoryCommitsPage'

import styles from './TreeCommits.module.scss'

interface Props {
    repo: TreePageRepositoryFields
    commitID: string
    filePath: string
    className?: string
}

const DEFAULT_FIRST = 10

/**
 * A list of commits in a tree (or in the entire repository for the root tree).
 */
export const TreeCommits: React.FunctionComponent<Props> = ({ repo, commitID, filePath, className }) => {
    const [showOlderCommits, setShowOlderCommits] = useState(false)
    const after = useMemo(
        () => (showOlderCommits ? null : formatISO(startOfDay(subYears(Date.now(), 1)))),
        [showOlderCommits]
    )

    const { connection, error, loading, hasNextPage, fetchMore, refetchAll } = useShowMorePagination<
        TreeCommitsResult,
        TreeCommitsVariables,
        GitCommitFields
    >({
        query: gql`
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
        `,
        variables: {
            repo: repo.id,
            revspec: commitID,
            first: DEFAULT_FIRST,
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
            fetchPolicy: 'cache-and-network',
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

    const showLinkToCommitsPage = connection && hasNextPage && connection.nodes.length > DEFAULT_FIRST

    return (
        <ConnectionContainer className={className}>
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
                        first={DEFAULT_FIRST}
                        connection={connection}
                        noun={showOlderCommits ? 'commit' : 'commit in the past year'}
                        pluralNoun={showOlderCommits ? 'commits' : 'commits in the past year'}
                        hasNextPage={hasNextPage}
                        emptyElement={null}
                    />
                    {hasNextPage ? (
                        showLinkToCommitsPage ? (
                            <Link to={`${repo.url}/-/commits${filePath ? `/${filePath}` : ''}`}>Show all commits</Link>
                        ) : (
                            <ShowMoreButton centered={true} onClick={fetchMore} />
                        )
                    ) : null}
                    {!hasNextPage && !showOlderCommits ? showAllCommits : null}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
