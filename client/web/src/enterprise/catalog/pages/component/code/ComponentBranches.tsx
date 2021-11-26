import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { useConnection } from '../../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
} from '../../../../../components/FilteredConnection/ui'
import { GitRefFields, ComponentBranchesResult, ComponentBranchesVariables } from '../../../../../graphql-operations'
import { gitReferenceFragments, GitReferenceNode } from '../../../../../repo/GitReference'

interface Props {
    component: Scalars['ID']
}

const COMPONENT_BRANCHES = gql`
    query ComponentBranches($component: ID!, $first: Int!, $withBehindAhead: Boolean = true) {
        node(id: $component) {
            __typename
            ... on Component {
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

export const ComponentBranches: React.FunctionComponent<Props> = ({ component }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        ComponentBranchesResult,
        ComponentBranchesVariables,
        GitRefFields
    >({
        query: COMPONENT_BRANCHES,
        variables: {
            component,
            first: FIRST,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || data.node.__typename !== 'Component') {
                throw new Error('not a component')
            }
            return data.node.branches
        },
    })
    return (
        <>
            <h4 className="sr-only">Active branches</h4>
            <ConnectionContainer>
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
