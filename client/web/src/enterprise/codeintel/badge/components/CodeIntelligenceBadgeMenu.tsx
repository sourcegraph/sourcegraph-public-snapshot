import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import BrainIcon from 'mdi-react/BrainIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import {
    Badge,
    Icon,
    LoadingSpinner,
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuList,
    Position,
} from '@sourcegraph/wildcard'

import { LSIFUploadState, LSIFIndexState } from '../../../../graphql-operations'
import {
    massageIndexerSupportMetadata,
    useCodeIntelStatus as defaultUseCodeIntelStatus,
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { CodeIntelligenceBadgeContentProps } from './CodeIntelligenceBadgeContent'
import { InternalCodeIntelligenceBadgeContent } from './InternalCodeIntelligenceBadgeContent'
import { UserFacingCodeIntelligenceBadgeContent } from './UserFacingCodeIntelligenceBadgeContent'

import styles from './CodeIntelligenceBadgeMenu.module.scss'

export const CodeIntelligenceBadgeMenu: React.FunctionComponent<
    React.PropsWithChildren<CodeIntelligenceBadgeContentProps>
> = ({
    isStorybook,
    now,
    useCodeIntelStatus = defaultUseCodeIntelStatus,
    useRequestedLanguageSupportQuery = defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery = defaultUseRequestLanguageSupportQuery,
    ...props
}) => {
    const { data, loading, error } = useCodeIntelStatus({
        variables: {
            repository: props.repoName,
            commit: props.revision,
            path: props.filePath,
        },
    })

    const indexerSupportMetadata = data && massageIndexerSupportMetadata(data)

    // Track when we render and then unmount the content component. We use the
    // onContentClose callback when the BadgeContent component below goes away.
    const [closed, setClosed] = useState<boolean | undefined>(undefined)
    const onContentClose = useCallback(() => setClosed(true), [setClosed])

    // Determine if we should show a badge CTA at all. The initial value will be
    // the current user's temporary setting (so we can show it until they interact).
    // This value may change on click of the top level menu.
    const [badgeUsed, setBadgeUsed] = useTemporarySetting('codeintel.badge.used', isStorybook === true)

    // Listen to all updates of badgeUsed and determine if we've ever seen a value that
    // was strictly false.
    const [badgeWasUnused, setBadgeWasUnused] = useState<boolean>(false)
    useEffect(() => setBadgeWasUnused(oldValue => oldValue || badgeUsed === false), [badgeUsed, setBadgeWasUnused])

    // The component is new to the current user if the badgeUsed value was ever strictly
    // false (temporary setting did not resolve to true), and we have not yet closed the
    // content section of the component.
    const isNew = badgeWasUnused && !closed

    const hasUploadErrors =
        (data?.recentUploads || [])
            .flatMap(uploads => uploads.uploads)
            .filter(upload => upload.state === LSIFUploadState.ERRORED).length > 0

    const hasIndexErrors =
        (data?.recentIndexes || [])
            .flatMap(indexes => indexes.indexes)
            .filter(index => index.state === LSIFIndexState.ERRORED).length > 0

    const needsAttention =
        indexerSupportMetadata &&
        (indexerSupportMetadata.indexerNames || []).filter(
            name =>
                // Count non-enabled and configurable indexers
                (indexerSupportMetadata.uploadsByIndexerName.get(name)?.length || 0) +
                    (indexerSupportMetadata.indexesByIndexerName.get(name)?.length || 0) ===
                    0 && indexerSupportMetadata.allIndexers.find(candidate => candidate.name === name) !== undefined
        ).length > 0

    const showDotError = hasUploadErrors || hasIndexErrors
    const showDotAttention = needsAttention || isNew
    const dotStyle = showDotError ? styles.braindotError : showDotAttention ? styles.braindotAttention : ''

    return (
        <Menu className="btn-icon">
            <>
                <MenuButton
                    className={classNames('text-decoration-none', styles.braindot, dotStyle)}
                    onClick={() => setBadgeUsed(true)}
                >
                    <Icon as={BrainIcon} />
                </MenuButton>

                <MenuList position={Position.bottomEnd} className={styles.dropdownMenu} isOpen={isStorybook}>
                    <MenuHeader>
                        Code intelligence{' '}
                        {isNew && (
                            <Badge variant="info" className="text-uppercase mx-2">
                                NEW
                            </Badge>
                        )}
                    </MenuHeader>

                    <MenuDivider />

                    {loading ? (
                        <div className="px-2 py-1">
                            <LoadingSpinner />
                        </div>
                    ) : error ? (
                        <div className="px-2 py-1">
                            <ErrorAlert prefix="Error loading repository summary" error={error} />
                        </div>
                    ) : data && indexerSupportMetadata ? (
                        <>
                            <UserFacingCodeIntelligenceBadgeContent
                                repoName={props.repoName}
                                indexerSupportMetadata={indexerSupportMetadata}
                                onClose={onContentClose}
                                now={now}
                                useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                                useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                                settingsCascade={props.settingsCascade}
                            />

                            <InternalCodeIntelligenceBadgeContent
                                data={data}
                                settingsCascade={props.settingsCascade}
                                now={now}
                            />
                        </>
                    ) : null}
                </MenuList>
            </>
        </Menu>
    )
}
