import * as React from 'react'

import { Timestamp } from '../../components/time/Timestamp'
import { MirrorRepositoryInfoFields } from '../../graphql-operations'

export const RepoUpdateSchedule: React.FunctionComponent<
    React.PropsWithChildren<{
        mirrorInfo: MirrorRepositoryInfoFields
    }>
> = ({ mirrorInfo }) => {
    const updateSchedule = mirrorInfo.updateSchedule
    const updateQueue = mirrorInfo.updateQueue

    return (
        <>
            {mirrorInfo.updatedAt === null ? (
                <>Not yet synced from code host.</>
            ) : (
                <>
                    Last synced <Timestamp date={mirrorInfo.updatedAt} />.
                </>
            )}{' '}
            {updateSchedule && (
                <>
                    Next scheduled sync <Timestamp date={updateSchedule.due} /> (position {updateSchedule.index + 1} out
                    of {updateSchedule.total} in the schedule).
                </>
            )}
            {updateQueue && !updateQueue.updating && (
                <>
                    Queued for sync (position {updateQueue.index + 1} out of {updateQueue.total} in the queue).
                </>
            )}
        </>
    )
}
