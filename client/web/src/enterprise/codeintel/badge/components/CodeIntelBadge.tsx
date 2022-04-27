import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import BrainIcon from 'mdi-react/BrainIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined } from '@sourcegraph/common'
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

import { CodeIntelBadgeProps as DefaultRepositoryMenuProps } from '../../../../codeintel/CodeIntelBadge'
import {
    LsifIndexFields,
    LsifUploadFields,
    PreciseSupportLevel,
    LSIFUploadState,
    LSIFIndexState,
    SearchBasedSupportLevel,
} from '../../../../graphql-operations'
import {
    useCodeIntelStatus as defaultUseCodeIntelStatus,
    UseCodeIntelStatusPayload,
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { IndexerSummary } from './IndexerSummary'
import { InternalBadgeContent } from './InternalBadgeContent'
import { Unsupported } from './Unsupported'

import styles from './CodeIntelBadge.module.scss'

export type CodeIntelBadgeProps = DefaultRepositoryMenuProps & {
    isOpen?: boolean
    now?: () => Date
    showBadgeCta?: boolean
    useCodeIntelStatus?: typeof defaultUseCodeIntelStatus
    useRequestedLanguageSupportQuery?: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery?: typeof defaultUseRequestLanguageSupportQuery
}

export const CodeIntelBadge: React.FunctionComponent<CodeIntelBadgeProps> = ({
    isOpen,
    now,
    showBadgeCta,
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
    // This value may change on click of the top level menu. The comparison with the
    // showBadgeCta prop is for use in storybooks (to set badgeUsed = true).
    const [badgeUsed, setBadgeUsed] = useTemporarySetting('codeintel.badge.used', showBadgeCta === false)

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

                <MenuList position={Position.bottomEnd} className={styles.dropdownMenu} isOpen={isOpen}>
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
                            <BadgeContent
                                repoName={props.repoName}
                                indexerSupportMetadata={indexerSupportMetadata}
                                now={now}
                                onClose={onContentClose}
                                useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                                useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                            />

                            <InternalBadgeContent data={data} settingsCascade={props.settingsCascade} now={now} />
                        </>
                    ) : null}
                </MenuList>
            </>
        </Menu>
    )
}

const BadgeContent: React.FunctionComponent<{
    repoName: string
    indexerSupportMetadata: IndexerSupportMetadata
    now?: () => Date
    onClose?: () => void
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
}> = ({
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
        <>
            {indexerNames.map((name, index) => (
                <React.Fragment key={`indexer-${name}`}>
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
                        useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                        useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                        now={now}
                    />
                </React.Fragment>
            ))}
        </>
    )
}

interface IndexerSupportMetadata {
    allIndexers: { name: string; url: string }[]
    indexerNames: string[]
    uploadsByIndexerName: Map<string, LsifUploadFields[]>
    indexesByIndexerName: Map<string, LsifIndexFields[]>
}

function massageIndexerSupportMetadata(data: UseCodeIntelStatusPayload): IndexerSupportMetadata {
    const allUploads = data.recentUploads.flatMap(uploads => uploads.uploads)
    const uploadsByIndexerName = groupBy<LsifUploadFields, string>(allUploads, getIndexerName)
    const allIndexes = data.recentIndexes.flatMap(indexes => indexes.indexes)
    const indexesByIndexerName = groupBy<LsifIndexFields, string>(allIndexes, getIndexerName)

    const nativelySupportedIndexers = (data.preciseSupport || [])
        .filter(support => support.supportLevel === PreciseSupportLevel.NATIVE)
        .map(support => support.indexers?.[0])
        .filter(isDefined)

    const allIndexers = [
        ...groupBy(
            [...allUploads, ...allIndexes]
                .map(index => index.indexer || undefined)
                .filter(isDefined)
                .concat(nativelySupportedIndexers),
            indexer => indexer.name
        ).values(),
    ].map(indexers => indexers[0])

    const languages = [
        ...new Set(
            data.searchBasedSupport
                ?.filter(support => support.supportLevel === SearchBasedSupportLevel.BASIC)
                .map(support => support.language)
        ),
    ].sort()
    const fakeIndexerNames = languages.map(name => `lsif-${name.toLowerCase()}`)
    const indexerNames = [...new Set(allIndexers.map(indexer => indexer.name).concat(fakeIndexerNames))].sort()

    return {
        allIndexers,
        indexerNames,
        uploadsByIndexerName,
        indexesByIndexerName,
    }
}

function groupBy<V, K>(values: V[], keyFn: (value: V) => K): Map<K, V[]> {
    return values.reduce(
        (map, value) => map.set(keyFn(value), (map.get(keyFn(value)) || []).concat([value])),
        new Map<K, V[]>()
    )
}

function getIndexerName(uploadOrIndexer: LsifUploadFields | LsifIndexFields): string {
    return uploadOrIndexer.indexer?.name || ''
}
