import * as React from 'react'

import { Accordion } from '@reach/accordion'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { Alert } from '@sourcegraph/wildcard'

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
        return <div className={styles.contents}>Loading...</div>
    }

    if (error) {
        logger.log(error)
        return (
            <div className={styles.contents}>
                <Alert variant="danger">Error getting ownership data.</Alert>
            </div>
        )
    }

    if (data?.node && data.node.__typename === 'Repository' && data.node.commit) {
        return (
            <Accordion
                as="table"
                collapsible={true}
                multiple={true}
                className={classNames(styles.table, styles.contents)}
            >
                <thead className="sr-only">
                    <tr>
                        <th>Show details</th>
                        <th>Contact</th>
                        <th>Owner</th>
                        <th>Reason</th>
                    </tr>
                </thead>
                {data.node.commit.blob?.ownership.nodes.map(ownership =>
                    ownership.owner.__typename === 'Person' ? (
                        <FileOwnershipEntry
                            key={ownership.owner.email}
                            person={ownership.owner}
                            reasons={ownership.reasons.filter(reason => reason.__typename === 'CodeownersFileEntry')}
                        />
                    ) : (
                        <></>
                    )
                )}
            </Accordion>
        )
    }

    return (
        <div className={styles.contents}>
            <Alert variant="info">No ownership data for this file.</Alert>
        </div>
    )
}

export const FETCH_OWNERS = gql`
    fragment OwnerFields on Person {
        email
        avatarURL
        displayName
        user {
            username
            displayName
            url
        }
    }

    fragment CodeownersFileEntryFields on CodeownersFileEntry {
        title
        description
    }

    query FetchOwnership($repo: ID!, $revision: String!, $currentPath: String!) {
        node(id: $repo) {
            ... on Repository {
                commit(rev: $revision) {
                    blob(path: $currentPath) {
                        ownership {
                            nodes {
                                owner {
                                    ...OwnerFields
                                }
                                reasons {
                                    ...CodeownersFileEntryFields
                                }
                            }
                        }
                    }
                }
            }
        }
    }
`
