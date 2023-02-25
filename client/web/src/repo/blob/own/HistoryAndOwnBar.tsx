import React, { useCallback } from 'react'

import { mdiAccount } from '@mdi/js'
import { useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Button, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

import { FetchOwnersAndHistoryResult, FetchOwnersAndHistoryVariables } from '../../../graphql-operations'
import { formatPersonName } from '../../../person/PersonLink'
import { GitCommitNode } from '../../commits/GitCommitNode'

import { FETCH_OWNERS_AND_HISTORY } from './grapqlQueries'

import styles from './HistoryAndOwnBar.module.scss'

export const HistoryAndOwnBar: React.FunctionComponent<{
    repoID: string
    revision?: string
    filePath: string
}> = ({ repoID, revision, filePath }) => {
    const navigate = useNavigate()

    const openOwnershipPanel = useCallback(() => {
        navigate({ hash: '#tab=ownership' })
    }, [navigate])

    const { data, loading } = useQuery<FetchOwnersAndHistoryResult, FetchOwnersAndHistoryVariables>(
        FETCH_OWNERS_AND_HISTORY,
        {
            variables: {
                repo: repoID,
                revision: revision ?? '',
                currentPath: filePath,
            },
        }
    )

    if (loading) {
        return (
            <div className={styles.wrapper}>
                <LoadingSpinner />
            </div>
        )
    }

    if (!(data?.node?.__typename === 'Repository' && data.node.commit)) {
        return <div className={styles.wrapper}>Error getting details about this file.</div>
    }

    const history = data?.node?.commit?.ancestors?.nodes?.[0]
    const ownership = data.node.commit?.blob?.ownership

    return (
        <div className={styles.wrapper}>
            {history && (
                <GitCommitNode
                    node={history}
                    extraCompact={true}
                    hideExpandCommitMessageBody={true}
                    className={styles.history}
                />
            )}
            {ownership && (
                <Tooltip content="Show ownership details" placement="left">
                    <Button className={styles.own} onClick={openOwnershipPanel}>
                        <div className={styles.ownBranding}>
                            <Icon svgPath={mdiAccount} aria-hidden="true" className={styles.ownIcon} /> Own
                        </div>

                        <div className={styles.ownItems}>
                            {ownership.nodes.slice(0, 2).map((ownership, index) => (
                                // There will only be 2 onwers max and they won't change, so
                                // it's safe to use the index as a key.
                                // eslint-disable-next-line react/no-array-index-key
                                <div className={styles.ownItem} key={index}>
                                    {ownership.owner.__typename === 'Person' ? (
                                        <>
                                            <UserAvatar user={ownership.owner} className="mx-2" />
                                            {formatPersonName(ownership.owner)}
                                        </>
                                    ) : (
                                        // TODO: Add support for teams.
                                        <></>
                                    )}
                                </div>
                            ))}
                            {ownership.totalCount > 2 && (
                                <div className={styles.ownMore}>+{ownership.totalCount - 2} more</div>
                            )}
                        </div>
                    </Button>
                </Tooltip>
            )}
        </div>
    )
}
