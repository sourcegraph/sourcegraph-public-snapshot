import classNames from 'classnames'
import * as H from 'history'
import React, { useState } from 'react'
import { useLocation } from 'react-router'
import { Link } from 'react-router-dom'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import { ConnectionSummary } from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import {
    GitCommitAncestorFields,
    RepositoryGitCommitResult,
    RepositoryGitCommitVariables,
} from '../../graphql-operations'

import { RevisionsPopoverTab } from './RevisionsPopoverTab'

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

const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({
    node,
    currentCommitID,
    location,
    getPathFromRevision,
    onClick,
}) => {
    const isCurrent = currentCommitID === node.oid
    return (
        <li key={node.oid} className="connection-popover__node revisions-popover-git-commit-node">
            <Link
                to={getPathFromRevision(location.pathname + location.search + location.hash, node.oid)}
                className={classNames(
                    'connection-popover__node-link',
                    isCurrent && 'connection-popover__node-link--active',
                    'revisions-popover-git-commit-node__link'
                )}
                onClick={onClick}
            >
                <code className="badge" title={node.oid}>
                    {node.abbreviatedOID}
                </code>
                <small className="revisions-popover-git-commit-node__message">{node.subject.slice(0, 200)}</small>
            </Link>
        </li>
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
}

const BATCH_COUNT = 15

export const RevisionsPopoverCommits: React.FunctionComponent<RevisionsPopoverCommitsProps> = ({
    repo,
    defaultBranch,
    getPathFromRevision,
    currentRev,
    noun,
    pluralNoun,
    currentCommitID,
    onSelect,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)
    const location = useLocation()

    const response = useConnection<RepositoryGitCommitResult, RepositoryGitCommitVariables, GitCommitAncestorFields>({
        query: REPOSITORY_GIT_COMMIT,
        variables: {
            query,
            first: BATCH_COUNT,
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
            first={BATCH_COUNT}
            noun={noun}
            pluralNoun={pluralNoun}
            hasNextPage={response.hasNextPage}
            connectionQuery={query}
        />
    )

    return (
        <RevisionsPopoverTab
            {...response}
            query={query}
            summary={summary}
            inputValue={searchValue}
            onInputChange={setSearchValue}
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
