import * as React from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'

import { Card, CardBody, CardHeader, Grid } from '@sourcegraph/wildcard'
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
        console.log(error)
        return <div>Error...</div>
    }

    if (data) {
        return (
            <Grid columnCount={4} className="mt-2">
                {data.node.commit.blob.ownership.map(
                    (own: FetchOwnershipResult['node']['commit']['blob']['ownership'][0]) => (
                        <Card>
                            <CardHeader>{own.owners.join(', ')}</CardHeader>
                            <CardBody>{own.reason}</CardBody>
                        </Card>
                    )
                )}
            </Grid>
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
