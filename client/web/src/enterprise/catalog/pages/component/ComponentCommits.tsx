import React from 'react'
import { Link } from 'react-router-dom'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import {
    GitCommitFields,
    SourceLocationSetCommitsResult,
    SourceLocationSetCommitsVariables,
} from '../../../../graphql-operations'
import { GitCommitNodeByline } from '../../../../repo/commits/GitCommitNodeByline'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'

interface Props {
    sourceLocationSet: Scalars['ID']
    className?: string
}

const SOURCE_LOCATION_SET_COMMITS = gql`
    query SourceLocationSetCommits($node: ID!, $first: Int!) {
        node(id: $node) {
            ... on SourceLocationSet {
                commits(first: $first) {
                    nodes {
                        ...GitCommitFields
                    }
                }
            }
        }
    }
    ${gitCommitFragment}
`

const FIRST = 10

export const SourceLocationSetCommits: React.FunctionComponent<Props> = ({ sourceLocationSet, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        SourceLocationSetCommitsResult,
        SourceLocationSetCommitsVariables,
        GitCommitFields
    >({
        query: SOURCE_LOCATION_SET_COMMITS,
        variables: {
            node: sourceLocationSet,
            first: FIRST,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-first',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || !('commits' in data.node) || !data.node.commits) {
                throw new Error('no commits associated with object')
            }
            return data.node.commits
        },
    })

    return (
        <>
            <h4 className="sr-only">Active branches</h4>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList as="ul" className="list-group list-group-flush">
                        {connection.nodes.map(commit => (
                            <GitCommit key={commit.oid} commit={commit} tag="li" className="list-group-item py-2" />
                        ))}
                    </ConnectionList>
                )}
                {loading && <ConnectionLoading className="my-2" />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={FIRST}
                            connection={connection}
                            noun="commit"
                            pluralNoun="commits"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No commits found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}

const GitCommit: React.FunctionComponent<{ commit: GitCommitFields; tag: 'li'; className?: string }> = ({
    commit,
    tag: Tag,
    className,
}) => (
    <Tag className={className}>
        <GitCommitNodeByline
            author={commit.author}
            committer={commit.committer}
            messageElement={
                <h4 className="h6 mb-0 text-truncate">
                    <Link to={commit.canonicalURL} className="text-body" title={commit.message}>
                        {commit.subject}
                    </Link>
                </h4>
            }
            className="d-flex align-items-center small text-muted"
        />
    </Tag>
)
