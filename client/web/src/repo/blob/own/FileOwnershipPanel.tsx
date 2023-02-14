import * as React from 'react'

import { Accordion } from '@reach/accordion'

import { logger } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'

import { FetchOwnershipResult, FetchOwnershipVariables } from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'

import styles from './FileOwnershipPanel.module.scss'

export const FileOwnershipPanel: React.FunctionComponent<
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
            <Accordion as="table" collapsible={true} multiple={true} className={styles.table}>
                <thead className="sr-only">
                    <tr>
                        <th>Show details</th>
                        <th>Contact</th>
                        <th>Owner</th>
                        <th>Reason</th>
                    </tr>
                </thead>
                {data.node.commit.blob?.ownership.map(own => (
                    <FileOwnershipEntry key={own.handle} person={own.person} reasons={own.reasons} />
                ))}
            </Accordion>
        )
    }

    return <div>No data</div>
}

export const FETCH_OWNERS = gql`
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
                                    avatarURL
                                    displayName
                                    user {
                                        username
                                        displayName
                                        url
                                    }
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
