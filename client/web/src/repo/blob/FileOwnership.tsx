import * as React from 'react'

import { mdiChat, mdiEmail } from '@mdi/js'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Button, Icon } from '@sourcegraph/wildcard'

import { FetchOwnershipResult, FetchOwnershipVariables } from '../../graphql-operations'

import { logger } from '@sourcegraph/common'
import styles from './FileOwnership.module.scss'

export const FileOwnership: React.FunctionComponent<
    React.PropsWithChildren<{
        repoID: string
        revision?: string
        filePath: string
    }>
> = props => {
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
        logger.log(error)
        return <div>Error...</div>
    }

    if (data?.node && data.node.__typename === 'Repository' && data.node.commit) {
        return (
            <table className={styles.table}>
                <thead className="sr-only">
                    <tr>
                        <th>Contact</th>
                        <th>Owner</th>
                        <th>Email</th>
                        <th>Reason</th>
                    </tr>
                </thead>
                <tbody>
                    {data.node.commit.blob?.ownership.map(own => (
                        <tr key={own.handle}>
                            <td>
                                <div className="d-flex">
                                    <Button variant="icon" className="mr-2">
                                        <Icon svgPath={mdiEmail} aria-label="email" />
                                    </Button>
                                    <Button variant="icon">
                                        <Icon svgPath={mdiChat} aria-label="chat" />
                                    </Button>
                                </div>
                            </td>
                            <td>{own.handle}</td>
                            <td>{own.person.email}</td>
                            <td>{own.reasons.join(', ')}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        )
    }

    return <div>No data</div>
}

const FETCH_OWNERS = gql`
    query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
        node(id: $repo) {
            ... on Repository {
                commit(rev: $revision) {
                    blob(path: $currentPath) {
                        ownership {
                            ... on Ownership {
                                handle
                                person {
                                    email
                                }
                                reasons
                            }
                        }
                    }
                }
            }
        }
    }
`
