import React from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { MenuDivider, Typography } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../../components/Collapsible'
import { Timestamp } from '../../../../components/time/Timestamp'
import { CodeIntelIndexer } from '../../shared/components/CodeIntelIndexer'
import { UseCodeIntelStatusPayload } from '../hooks/useCodeIntelStatus'

import { UploadOrIndexMetaTable } from './UploadOrIndexMetaTable'

export type InternalCodeIntelligenceBadgeContentProps = SettingsCascadeProps & {
    data: UseCodeIntelStatusPayload
    now?: () => Date
}

export const InternalCodeIntelligenceBadgeContent: React.FunctionComponent<
    React.PropsWithChildren<InternalCodeIntelligenceBadgeContentProps>
> = ({ data, now, settingsCascade }) => {
    const forNerds =
        !isErrorLike(settingsCascade.final) &&
        settingsCascade.final?.experimentalFeatures?.codeIntelRepositoryBadge?.forNerds

    if (!forNerds) {
        return null
    }

    const preciseSupportLevels = [...new Set((data.preciseSupport || []).map(support => support.supportLevel))].sort()
    const searchBasedSupportLevels = [
        ...new Set((data?.searchBasedSupport || []).map(support => support.supportLevel)),
    ].sort()

    return (
        <>
            <MenuDivider />

            <div className="px-2 py-1">
                <Collapsible titleAtStart={true} title={<Typography.H3>Activity (repo)</Typography.H3>}>
                    <div>
                        <span>
                            Last auto-indexing job schedule attempt:{' '}
                            {data.lastIndexScan ? <Timestamp date={data.lastIndexScan} now={now} /> : <>never</>}
                        </span>
                    </div>
                    <div>
                        <span>
                            Last upload retention scan:{' '}
                            {data.lastUploadRetentionScan ? (
                                <Timestamp date={data.lastUploadRetentionScan} now={now} />
                            ) : (
                                <>never</>
                            )}
                        </span>
                    </div>
                </Collapsible>

                <Collapsible titleAtStart={true} title={<Typography.H3>Support (tree)</Typography.H3>}>
                    <ul>
                        {preciseSupportLevels.map(supportLevel => (
                            <li key={`precise-support-level-${supportLevel}`}>
                                <code>{supportLevel}</code>
                                <ul>
                                    {data.preciseSupport
                                        ?.filter(support => support.supportLevel === supportLevel)
                                        .map(support =>
                                            support.indexers?.map(indexer => (
                                                <li key={`precise-support-level-${supportLevel}-${indexer.name}`}>
                                                    <CodeIntelIndexer indexer={indexer} /> (
                                                    {support.confidence && (
                                                        <span className="text-muted">{support.confidence}</span>
                                                    )}
                                                    )
                                                </li>
                                            ))
                                        )}
                                </ul>
                            </li>
                        ))}

                        {searchBasedSupportLevels.map(supportLevel => (
                            <li key={`search-support-level-${supportLevel}`}>
                                <code>{supportLevel}</code>
                                <ul>
                                    {data.searchBasedSupport
                                        ?.filter(support => support.supportLevel === supportLevel)
                                        .map(support => (
                                            <li key={`search-support-level-${supportLevel}-${support.language}`}>
                                                {support.language}
                                            </li>
                                        ))}
                                </ul>
                            </li>
                        ))}
                    </ul>
                </Collapsible>

                <Collapsible titleAtStart={true} title={<Typography.H3>Recent uploads (repo)</Typography.H3>}>
                    <UploadOrIndexMetaTable
                        prefix="recent-uploads"
                        nodes={data.recentUploads.flatMap(namespacedUploads => namespacedUploads.uploads)}
                    />
                </Collapsible>

                <Collapsible titleAtStart={true} title={<Typography.H3>Recent indexes (repo)</Typography.H3>}>
                    <UploadOrIndexMetaTable
                        prefix="recent-indexes"
                        nodes={data.recentIndexes.flatMap(namespacedIndexes => namespacedIndexes.indexes)}
                    />
                </Collapsible>

                <Collapsible titleAtStart={true} title={<Typography.H3>Uploads providing intel (tree)</Typography.H3>}>
                    <UploadOrIndexMetaTable prefix="active-uploads" nodes={data.activeUploads} />
                </Collapsible>
            </div>
        </>
    )
}
