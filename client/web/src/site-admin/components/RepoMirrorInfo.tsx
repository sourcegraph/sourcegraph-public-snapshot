import * as React from 'react'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Code, Text, Tooltip } from '@sourcegraph/wildcard'

import type { MirrorRepositoryInfoFields } from '../../graphql-operations'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

export const RepoMirrorInfo: React.FunctionComponent<
    React.PropsWithChildren<{
        mirrorInfo: MirrorRepositoryInfoFields
    }>
> = ({ mirrorInfo }) => (
    <>
        <Text className="mb-0 text-muted">
            <small>
                {mirrorInfo.updatedAt === null ? (
                    <>Not yet synced from code host.</>
                ) : (
                    <>
                        Last synced <Timestamp date={mirrorInfo.updatedAt} />. Next sync time:{' '}
                        {mirrorInfo.nextSyncAt === null ? (
                            <>No update scheduled</>
                        ) : (
                            <Timestamp date={mirrorInfo.nextSyncAt} />
                        )}
                        . Size: {prettyBytesBigint(BigInt(mirrorInfo.byteSize))}.
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
                        {mirrorInfo.cloneInProgress && (mirrorInfo.cloneProgress ?? '').trim() !== '' ? (
                            <>
                                <br />
                                <Code>{mirrorInfo.cloneProgress}</Code>
                            </>
                        ) : (
                            ''
                        )}
                    </>
                )}
            </small>
        </Text>
    </>
)
