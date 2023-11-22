import type { FC } from 'react'

import { mdiFile } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Icon, Link, ErrorAlert } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionLoading,
    ShowMoreButton,
    SummaryContainer,
} from '../components/FilteredConnection/ui'
import type { FetchCommitsResult, FetchCommitsVariables, GitCommitFields, Scalars } from '../graphql-operations'
import { replaceRevisionInURL } from '../util/url'

import { GitCommitNode } from './commits/GitCommitNode'
import { gitCommitFragment } from './commits/RepositoryCommitsPage'

import styles from './RepoRevisionSidebarCommits.module.scss'

interface CommitNodeProps {
    node: GitCommitFields
    preferAbsoluteTimestamps: boolean
}

const CommitNode: FC<CommitNodeProps> = ({ node, preferAbsoluteTimestamps }) => {
    const location = useLocation()

    return (
        <li className={classNames(styles.commitContainer, 'list-group-item p-0')}>
            <GitCommitNode
                className={styles.commitNode}
                compact={true}
                node={node}
                hideExpandCommitMessageBody={true}
                preferAbsoluteTimestamps={preferAbsoluteTimestamps}
                afterElement={
                    <Link
                        to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.oid)}
                        className={classNames(styles.fileIcon, 'ml-2')}
                        title="View current file at this commit"
                    >
                        <Icon aria-hidden={true} svgPath={mdiFile} />
                    </Link>
                }
            />
        </li>
    )
}

interface Props extends Partial<RevisionSpec>, FileSpec {
    repoID: Scalars['ID']
    preferAbsoluteTimestamps: boolean
    defaultPageSize?: number
}

export const RepoRevisionSidebarCommits: FC<Props> = props => {
    const { connection, error, loading, hasNextPage, fetchMore } = useShowMorePagination<
        FetchCommitsResult,
        FetchCommitsVariables,
        GitCommitFields
    >({
        query: FETCH_COMMITS,
        variables: {
            afterCursor: null,
            first: props.defaultPageSize || 100,
            query: '',
            repo: props.repoID,
            revision: props.revision || '',
            currentPath: props.filePath || '',
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                return { nodes: [] }
            }
            if (node.__typename !== 'Repository') {
                return { nodes: [] }
            }
            if (!node.commit?.ancestors?.nodes) {
                return { nodes: [] }
            }

            return node.commit.ancestors
        },
        options: {
            // Currently "after" is used as a commit filtering option to return
            // commits after a specific date. Currently the pagination is
            // implemented by using afterCursor instead and setting this boolean
            // will ensure that the pagination works correctly.
            useAlternateAfterCursor: true,
            fetchPolicy: 'cache-first',
        },
    })

    return (
        <ConnectionContainer>
            {error && <ErrorAlert error={error} />}
            {connection?.nodes.map(node => (
                <CommitNode key={node.id} node={node} preferAbsoluteTimestamps={props.preferAbsoluteTimestamps} />
            ))}
            {loading && <ConnectionLoading />}
            {!loading && connection && (
                <SummaryContainer centered={true}>
                    {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}

const FETCH_COMMITS = gql`
    query FetchCommits(
        $repo: ID!
        $revision: String!
        $first: Int
        $currentPath: String
        $query: String
        $afterCursor: String
    ) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    ancestors(first: $first, query: $query, path: $currentPath, afterCursor: $afterCursor) {
                        ...CommitAncestorsConnectionFields
                    }
                }
            }
        }
    }

    ${gitCommitFragment}

    fragment CommitAncestorsConnectionFields on GitCommitConnection {
        nodes {
            ...GitCommitFields
        }
        pageInfo {
            endCursor
            hasNextPage
        }
    }
`
