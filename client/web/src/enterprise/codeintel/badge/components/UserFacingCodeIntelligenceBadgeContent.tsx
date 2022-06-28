import React, { useEffect } from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { MenuDivider } from '@sourcegraph/wildcard'

import {
    IndexerSupportMetadata,
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { IndexerSummary } from './IndexerSummary'
import { Unsupported } from './Unsupported'

import styles from './UserFacingCodeIntelligenceBadgeContent.module.scss'

export interface CodeIntelligenceBadgeContentStorybookProps {
    now?: () => Date
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
}

export interface UserFacingCodeIntelligenceBadgeContentProps
    extends SettingsCascadeProps,
        CodeIntelligenceBadgeContentStorybookProps {
    repoName: string
    indexerSupportMetadata: IndexerSupportMetadata
    onClose?: () => void
}

export const UserFacingCodeIntelligenceBadgeContent: React.FunctionComponent<
    React.PropsWithChildren<UserFacingCodeIntelligenceBadgeContentProps>
> = ({
    repoName,
    indexerSupportMetadata: { allIndexers, indexerNames, uploadsByIndexerName, indexesByIndexerName },
    onClose,
    now,
    useRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery,
}) => {
    // Call onClose when this component unmounts
    useEffect(() => onClose, [onClose])

    // Expand badges to be as large as the maximum badge when we are displaying
    // badges of different types. This condition checks that there's at least
    // two distinct states being displayed in the following rendered component.
    const className =
        new Set(
            indexerNames.map(name =>
                (uploadsByIndexerName.get(name)?.length || 0) + (indexesByIndexerName.get(name)?.length || 0) > 0
                    ? 'enabled'
                    : allIndexers.find(candidate => candidate.name === name) !== undefined
                    ? 'configurable'
                    : 'unavailable'
            )
        ).size > 1
            ? styles.badgeMultiple
            : undefined

    return indexerNames.length === 0 ? (
        <Unsupported />
    ) : (
        <ul className={styles.list}>
            {indexerNames.map((name, index) => (
                <li key={`indexer-${name}`}>
                    {index > 0 && <MenuDivider />}
                    <IndexerSummary
                        repoName={repoName}
                        summary={{
                            name,
                            uploads: uploadsByIndexerName.get(name) || [],
                            indexes: indexesByIndexerName.get(name) || [],
                            indexer: allIndexers.find(candidate => candidate.name === name),
                        }}
                        className={className}
                        now={now}
                        useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                        useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                    />
                </li>
            ))}
        </ul>
    )
}
