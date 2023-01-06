import * as React from 'react'

import { Accordion } from '@reach/accordion'

import { logger } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'

import { FetchOwnershipResult, FetchOwnershipVariables } from '../../graphql-operations'

import { FileOwnershipReasons } from './FileOwnershipReasons'

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
                <Accordion as="tbody" collapsible={true} multiple={true}>
                    {data.node.commit.blob?.ownership.map(own => (
                        <FileOwnershipReasons
                            key={own.handle}
                            email={own.person.email}
                            handle={own.handle}
                            reasons={own.reasons}
                        />
                    ))}
                </Accordion>
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
                                reasons {
                                    ... on CodeownersFileEntry {
                                        title
                                        description
                                    }
                                    ... on RecentContributor {
                                        title
                                        description
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
`
