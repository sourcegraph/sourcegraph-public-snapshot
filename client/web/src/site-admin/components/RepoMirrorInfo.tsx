import * as React from 'react'

import { mdiCloudOutline } from '@mdi/js'

import { Icon, LoadingSpinner, Text, Tooltip } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { MirrorRepositoryInfoFields } from '../../graphql-operations'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

export const RepoMirrorInfo: React.FunctionComponent<
    React.PropsWithChildren<{
        mirrorInfo: MirrorRepositoryInfoFields
    }>
> = ({ mirrorInfo }) => (
    <>
        {mirrorInfo.cloneInProgress && (
            <small className="ml-2 text-success">
                <LoadingSpinner /> Cloning
            </small>
        )}
        {!mirrorInfo.cloneInProgress && !mirrorInfo.cloned && (
            <Tooltip content="Visit the repository to clone it. See its mirroring settings for diagnostics.">
                <small className="ml-2 text-muted">
                    <Icon aria-hidden={true} svgPath={mdiCloudOutline} /> Not yet cloned
                </small>
            </Tooltip>
        )}
        <Text className="mb-0 text-muted">
            <small>
                {mirrorInfo.updatedAt === null ? (
                    <>Not yet synced from code host.</>
                ) : (
                    <>
                        Last synced <Timestamp date={mirrorInfo.updatedAt} />. Size:{' '}
                        {prettyBytesBigint(BigInt(mirrorInfo.byteSize))}.
                        {mirrorInfo.shard !== null && <> Shard: {mirrorInfo.shard}</>}
                        {mirrorInfo.shard === null && (
                            <>
                                {' '}
                                Shard:{' '}
                                <Tooltip content="The repo has not yet been picked up by a gitserver instance.">
                                    <span>not assigned</span>
                                </Tooltip>
                            </>
                        )}
                    </>
                )}
            </small>
        </Text>
    </>
)
