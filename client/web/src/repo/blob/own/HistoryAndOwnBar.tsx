import React, { useCallback, useEffect } from 'react'

import { mdiAccount } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Alert, Button, Icon, LoadingSpinner, Tooltip } from '@sourcegraph/wildcard'

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

    const { data, error, loading } = useQuery<FetchOwnersAndHistoryResult, FetchOwnersAndHistoryVariables>(
        FETCH_OWNERS_AND_HISTORY,
        {
            variables: {
                repo: repoID,
                revision: revision ?? '',
                currentPath: filePath,
            },
        }
    )

    useEffect(() => {
        if (error) {
            logger.error(error)
        }
    }, [error])

    if (loading) {
        return (
            <div className={styles.wrapper}>
                <LoadingSpinner />
            </div>
        )
    }

    if (error || !(data?.node?.__typename === 'Repository' && data.node.commit)) {
        return (
            <div className={styles.wrapper}>
                <Alert variant="danger" className="mb-0 py-1" aria-live="polite">
                    Error getting history and ownership details about this file.
                </Alert>
            </div>
        )
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
                            {ownership.nodes.length === 0 && (
                                <div className={classNames(styles.ownItem, styles.ownItemEmpty)}>No owner found</div>
                            )}

                            {ownership.nodes.slice(0, 2).map((ownership, index) => (
                                // There will only be 2 owners max and they won't change, so
                                // it's safe to use the index as a key.
                                // eslint-disable-next-line react/no-array-index-key
                                <div className={styles.ownItem} key={index}>
                                    {ownership.owner.__typename === 'Person' && (
                                        <>
                                            <UserAvatar user={ownership.owner} className="mx-2" inline={true} />
                                            {formatPersonName(ownership.owner)}
                                        </>
                                    )}
                                    {ownership.owner.__typename === 'Team' && (
                                        <>
                                            <TeamAvatar
                                                team={{
                                                    ...ownership.owner,
                                                    displayName: ownership.owner.teamDisplayName,
                                                }}
                                                className="mx-2"
                                                inline={true}
                                            />
                                            {ownership.owner.teamDisplayName || ownership.owner.name}
                                        </>
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
