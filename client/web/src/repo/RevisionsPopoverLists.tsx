import React from 'react'
import { useLocation } from 'react-router'

import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import {
    GitRefConnectionFields,
    GitRefFields,
    RepositoryGitRefsResult,
    RepositoryGitRefsVariables,
} from '../graphql-operations'

import { REPOSITORY_GIT_REFS } from './GitReference'
import { GitReferencePopoverNode } from './RevisionsPopover'

interface BranchesRevisionsListProps {
    type: GitRefType
    query: string
    repo: Scalars['ID']
    defaultBranch: string
    getURLFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined
}

export const ReferencesRevisionsList: React.FunctionComponent<BranchesRevisionsListProps> = ({
    type,
    query,
    repo,
    defaultBranch,
    getURLFromRevision,
    currentRev,
    noun,
    pluralNoun,
}) => {
    const location = useLocation()
    const { connection, loading, errors, hasNextPage, fetchMore } = useConnection<
        RepositoryGitRefsResult,
        RepositoryGitRefsVariables,
        GitRefFields
    >({
        query: REPOSITORY_GIT_REFS,
        variables: {
            query,
            first: 50,
            repo,
            type,
            withBehindAhead: false,
        },
        getConnection: ({ data, errors }) => {
            if (!data || !data.node || !data.node.gitRefs) {
                throw createAggregateError(errors)
            }
            return data.node.gitRefs
        },
    })

    if (loading) {
        console.log('Loading...')
    }

    if (errors) {
        console.log('Got errors', errors)
    }

    if (connection) {
        console.log('Got connection!')
        console.log(connection.nodes)
    }

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
        <>
            <SummaryContainer>{query && summary}</SummaryContainer>
            <ConnectionList className="connection-popover__nodes">
                {connection?.nodes?.map((node, index) => (
                    <GitReferencePopoverNode
                        key={index}
                        node={node}
                        currentRevision={currentRev}
                        defaultBranch={defaultBranch}
                        getURLFromRevision={getURLFromRevision}
                        location={location}
                    />
                ))}
            </ConnectionList>
            {loading && <ConnectionLoading />}
            {connection && (
                <SummaryContainer>
                    <h1>yep</h1>
                    {summary}
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </>
    )
}
