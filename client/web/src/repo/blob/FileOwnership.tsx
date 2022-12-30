import * as React from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'

import { FetchOwnershipResult, FetchOwnershipVariables } from '../../graphql-operations'

export const FileOwnership: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const { data, loading, error } = useQuery<FetchOwnershipResult, FetchOwnershipVariables>(FETCH_OWNERS, {
        variables: {
            repo: props.repoID,
            revision: props.revision ?? '',
            currentPath: props.filePath,
        },
    })
    if (loading) {
        return <div>Loading...</div>
    }

    if (error) {
        return <div>Error...</div>
    }

    if (data) {
        return data.node.commit.blob.ownership.map(
            (own: FetchOwnershipResult['node']['commit']['blob']['ownership'][0]) => (
                <>
                    <div>{own.owners.join(', ')}</div>
                    <div>{own.reason}</div>
                </>
            )
        )
    }

    return <div>No data</div>
}

const FETCH_OWNERS = gql`
    query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revision) {
                    blob(path: $currentPath) {
                        ownership {
                            ... on Owner {
                                owners
                                reason
                            }
                        }
                    }
                }
            }
        }
    }
`
