import React, { useMemo, useState } from 'react'

import type * as H from 'history'
import { useLocation } from 'react-router-dom'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Badge, useDebounce, useWindowSize } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { ConnectionSummary } from '../../components/FilteredConnection/ui'
import type {
    GitCommitAncestorFields,
    RepositoryGitCommitResult,
    RepositoryGitCommitVariables,
} from '../../graphql-operations'

import { ConnectionPopoverNode, ConnectionPopoverNodeLink } from './components'
import { RevisionsPopoverTab } from './RevisionsPopoverTab'
import { getBatchCount } from './utils'

import styles from './RevisionsPopoverCommits.module.scss'

export const REPOSITORY_GIT_COMMIT = gql`
    query RepositoryGitCommit($repo: ID!, $first: Int, $revision: String!, $query: String) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    __typename
                    ancestors(first: $first, query: $query) {
                        __typename
                        ...GitCommitAncestorsConnectionFields
                    }
                }
            }
        }
    }

    fragment GitCommitAncestorsConnectionFields on GitCommitConnection {
        nodes {
            __typename
            ...GitCommitAncestorFields
        }
        pageInfo {
            hasNextPage
        }
    }

    fragment GitCommitAncestorFields on GitCommit {
        id
        oid
        abbreviatedOID
        author {
            person {
                name
                avatarURL
            }
            date
        }
        subject
    }
`

interface GitCommitNodeProps {
    node: GitCommitAncestorFields

    currentCommitID: string | undefined

    location: H.Location

    getPathFromRevision: (href: string, revision: string) => string

    onClick?: React.MouseEventHandler<HTMLAnchorElement>
}

const GitCommitNode: React.FunctionComponent<React.PropsWithChildren<GitCommitNodeProps>> = ({
    node,
    currentCommitID,
    location,
    getPathFromRevision,
    onClick,
}) => {
    const isCurrent = currentCommitID === node.oid
    return (
        <ConnectionPopoverNode key={node.oid}>
            <ConnectionPopoverNodeLink
                to={getPathFromRevision(location.pathname + location.search + location.hash, node.oid)}
                active={isCurrent}
                onClick={onClick}
            >
                <Badge title={node.oid} className="px-1 py-0" as="code">
                    {node.abbreviatedOID}
                </Badge>
                <small title={node.author.date} className={styles.message}>
                    {node.subject.slice(0, 200)}
                </small>
            </ConnectionPopoverNodeLink>
        </ConnectionPopoverNode>
    )
}

interface RevisionsPopoverCommitsProps {
    repo: Scalars['ID']
    defaultBranch: string
    getPathFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    onSelect?: (node: GitCommitAncestorFields) => void

    tabLabel: string
}

export const RevisionsPopoverCommits: React.FunctionComponent<
    React.PropsWithChildren<RevisionsPopoverCommitsProps>
> = ({
    repo,
    defaultBranch,
    getPathFromRevision,
    currentRev,
    noun,
    pluralNoun,
    currentCommitID,
    onSelect,
    tabLabel,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)
    const location = useLocation()
    const { height } = useWindowSize()
    const [batchCount, setBatchCount] = useState(getBatchCount(height))

    /*
     * Set batch count based on current screen height
     * to avoid jumping tabs in the UI.
     * */
    useMemo(() => {
        setBatchCount(getBatchCount(height))
    }, [height])

    const response = useShowMorePagination<
        RepositoryGitCommitResult,
        RepositoryGitCommitVariables,
        GitCommitAncestorFields
    >({
        query: REPOSITORY_GIT_COMMIT,
        variables: {
            query,
            first: batchCount,
            repo,
            revision: currentRev || defaultBranch,
        },
        getConnection: response => {
            const { node } = dataOrThrowErrors(response)

            if (!node) {
                throw new Error(`Repository ${repo} not found`)
            }

            if (node.__typename !== 'Repository') {
                throw new Error(`Node is a ${node.__typename}, not a Repository`)
            }

            if (!node.commit) {
                // Did not find a commit for the current revision, the user may have provided an invalid revision.
                // Avoid erroring here so this can be reflected correctly in the UI.
                return {
                    nodes: [],
                }
            }

            if (!node.commit.ancestors) {
                throw new Error(`Cannot load ancestors for repository ${repo}`)
            }

            return node.commit.ancestors
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    const summary = response.connection && (
        <ConnectionSummary
            connection={response.connection}
            first={batchCount}
            noun={noun}
            pluralNoun={pluralNoun}
            hasNextPage={response.hasNextPage}
            connectionQuery={query}
            compact={true}
        />
    )

    return (
        <RevisionsPopoverTab
            {...response}
            query={query}
            summary={summary}
            inputValue={searchValue}
            onInputChange={setSearchValue}
            inputAriaLabel={tabLabel}
        >
            {response.connection?.nodes?.map((node, index) => (
                <GitCommitNode
                    key={index}
                    node={node}
                    currentCommitID={currentCommitID}
                    location={location}
                    getPathFromRevision={getPathFromRevision}
                    onClick={() => onSelect?.(node)}
                />
            ))}
        </RevisionsPopoverTab>
    )
}
