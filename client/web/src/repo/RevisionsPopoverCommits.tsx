import classNames from 'classnames'
import * as H from 'history'
import React, { useState } from 'react'
import { useLocation } from 'react-router'
import { Link } from 'react-router-dom'

import { CircleChevronLeftIcon } from '@sourcegraph/shared/src/components/icons'
import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import { GitCommitAncestorFields, RepositoryGitCommitResult, RepositoryGitCommitVariables } from '../graphql-operations'

export const REPOSITORY_GIT_COMMIT = gql`
    query RepositoryGitCommit($repo: ID!, $first: Int, $revision: String!, $query: String) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    ancestors(first: $first, query: $query) {
                        ...GitCommitAncestorsConnectionFields
                    }
                }
            }
        }
    }

    fragment GitCommitAncestorsConnectionFields on GitCommitConnection {
        nodes {
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

    getURLFromRevision: (href: string, revision: string) => string
}

const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({
    node,
    currentCommitID,
    location,
    getURLFromRevision,
}) => {
    const isCurrent = currentCommitID === node.oid
    return (
        <li key={node.oid} className="connection-popover__node revisions-popover-git-commit-node">
            <Link
                to={getURLFromRevision(location.pathname + location.search + location.hash, node.oid)}
                className={classNames(
                    'connection-popover__node-link',
                    isCurrent && 'connection-popover__node-link--active',
                    'revisions-popover-git-commit-node__link'
                )}
            >
                <code className="revisions-popover-git-commit-node__oid" title={node.oid}>
                    {node.abbreviatedOID}
                </code>
                <small className="revisions-popover-git-commit-node__message">{node.subject.slice(0, 200)}</small>
            </Link>
        </li>
    )
}

interface RevisionCommitsTabProps {
    repo: Scalars['ID']
    defaultBranch: string
    getURLFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string
}

export const RevisionCommitsTab: React.FunctionComponent<RevisionCommitsTabProps> = ({
    repo,
    defaultBranch,
    getURLFromRevision,
    currentRev,
    noun,
    pluralNoun,
    currentCommitID,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const debouncedSearchValue = useDebounce(searchValue, 200)
    const location = useLocation()

    const { connection, loading, errors, hasNextPage, fetchMore } = useConnection<
        RepositoryGitCommitResult,
        RepositoryGitCommitVariables,
        GitCommitAncestorFields
    >({
        query: REPOSITORY_GIT_COMMIT,
        variables: {
            query: debouncedSearchValue,
            first: 15,
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

            if (!node.commit?.ancestors) {
                throw new Error(`Cannot load ancestors for repository ${repo}`)
            }

            return node.commit.ancestors
        },
    })

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            noun={noun}
            pluralNoun={pluralNoun}
            totalCount={connection.totalCount ?? null}
            hasNextPage={hasNextPage}
        />
    )

    return (
        <ConnectionContainer compact={true} className="connection-popover__content">
            <ConnectionForm
                inputValue={searchValue}
                onInputChange={event => setSearchValue(event.target.value)}
                autoFocus={true}
                inputPlaceholder="Find..."
                inputClassName="connection-popover__input"
            />
            <SummaryContainer>{searchValue && summary}</SummaryContainer>
            <ConnectionList className="connection-popover__nodes">
                {connection?.nodes?.map((node, index) => (
                    <GitCommitNode
                        key={index}
                        node={node}
                        currentCommitID={currentCommitID}
                        location={location}
                        getURLFromRevision={getURLFromRevision}
                    />
                ))}
            </ConnectionList>
            {loading && <ConnectionLoading />}
            {connection && (
                <SummaryContainer>
                    {summary}
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
