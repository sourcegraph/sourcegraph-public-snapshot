import React, { useMemo } from 'react'

import classNames from 'classnames'
import { sortBy } from 'lodash'

import { Badge } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

interface RepoMetadataItemProps {
    metadataKey: string
    metadataValue: string | null | undefined
    small?: boolean
}

const RepoMetadataItem: React.FunctionComponent<RepoMetadataItemProps> = ({ metadataKey, metadataValue, small }) => (
    <div className="d-flex align-items-stretch justify-content-center">
        <Badge
            as="dt"
            small={small}
            variant="info"
            aria-label={`Repository metadata key "${metadataKey}"`}
            className={classNames('m-0', { [styles.roundedRightNone]: metadataValue })}
        >
            {metadataKey}
        </Badge>
        <Badge
            as="dd"
            small={small}
            variant="secondary"
            aria-label={`Repository metadata value "${metadataValue}" for key=${metadataKey}`}
            aria-hidden={!metadataValue}
            className={classNames(styles.roundedLeftNone, 'm-0', {
                ['d-none']: !metadataValue,
            })}
        >
            {metadataValue}
        </Badge>
    </div>
)

interface RepoMetadataProps {
    metadata: [string, string | null | undefined][]
    className?: string
    small?: boolean
}

export const RepoMetadata: React.FunctionComponent<RepoMetadataProps> = ({ metadata, small, className }) => {
    const sortedPairs = useMemo(() => sortBy(metadata), [metadata])
    if (sortedPairs.length === 0) {
        return null
    }
    return (
        <dl className={classNames(styles.repoMetadata, 'd-flex align-items-start flex-wrap m-0', className)}>
            {sortedPairs.map(([key, value]) => (
                <RepoMetadataItem key={`${key}:${value}`} metadataKey={key} metadataValue={value} small={small} />
            ))}
        </dl>
    )
}
