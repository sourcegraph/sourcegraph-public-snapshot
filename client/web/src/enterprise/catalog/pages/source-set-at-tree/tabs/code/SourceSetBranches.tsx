import React from 'react'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import { GitRefFields, SourceSetBranchesResult, SourceSetBranchesVariables } from '../../../../../../graphql-operations'
import { gitReferenceFragments, GitReferenceNode } from '../../../../../../repo/GitReference'

interface Props {
    sourceSet: Scalars['ID']
    className?: string
}

const SOURCE_SET_BRANCHES = gql`
    query SourceSetBranches($node: ID!, $first: Int!, $withBehindAhead: Boolean = true) {
        node(id: $node) {
            ... on SourceSet {
                branches(first: $first) {
                    nodes {
                        ...GitRefFields
                        repository {
                            url
                        }
                    }
                }
            }
        }
    }
    ${gitReferenceFragments}
`

const FIRST = 50

export const SourceSetBranches: React.FunctionComponent<Props> = ({ sourceSet, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        SourceSetBranchesResult,
        SourceSetBranchesVariables,
        GitRefFields
    >({
        query: SOURCE_SET_BRANCHES,
        variables: {
            node: sourceSet,
            first: FIRST,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-first',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || !('branches' in data.node) || !data.node.branches) {
                throw new Error('no branches associated with object')
            }
            return data.node.branches
        },
    })
    return (
        <>
            <h4 className="sr-only">Active branches</h4>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList as="div">
                        {connection?.nodes?.map(branch => (
                            <GitReferenceNode
                                key={branch.id}
                                node={branch}
                                url={`${branch.repository.url}/-/compare/...${encodeURIComponent(branch.abbrevName)}`}
                            />
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
                            noun="branch"
                            pluralNoun="branches"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No branches found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}
