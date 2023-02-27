import * as React from 'react'

import { Accordion } from '@reach/accordion'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Alert } from '@sourcegraph/wildcard'

import { FetchOwnershipResult, FetchOwnershipVariables } from '../../../graphql-operations'

import { FileOwnershipEntry } from './FileOwnershipEntry'
import { FETCH_OWNERS } from './grapqlQueries'

import styles from './FileOwnershipPanel.module.scss'

export const FileOwnershipPanel: React.FunctionComponent<{
    repoID: string
    revision?: string
    filePath: string
}> = ({ repoID, revision, filePath }) => {
    const { data, loading, error } = useQuery<FetchOwnershipResult, FetchOwnershipVariables>(FETCH_OWNERS, {
        variables: {
            repo: repoID,
            revision: revision ?? '',
            currentPath: filePath,
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

    if (
        data?.node &&
        data.node.__typename === 'Repository' &&
        data.node.commit?.blob &&
        data.node.commit.blob.ownership.nodes.length > 0
    ) {
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
                        // TODO #48303: Add support for teams.
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
