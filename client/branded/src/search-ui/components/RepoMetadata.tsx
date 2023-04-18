import React, { useMemo } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { sortBy } from 'lodash'

import { Badge, Button, Icon } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

interface RepoMetadataItemProps {
    metadataKey: string
    metadataValue: string | null | undefined
    small?: boolean
    onDelete?: () => void
}

const RepoMetadataItem: React.FunctionComponent<RepoMetadataItemProps> = ({
    metadataKey,
    metadataValue,
    small,
    onDelete,
}) => (
    <div className="d-flex align-items-stretch justify-content-center">
        <Badge
            as="dt"
            small={small}
            variant="info"
            aria-label={`Repository metadata key "${metadataKey}"`}
            className={classNames('m-0', { [styles.roundedRightNone]: metadataValue || onDelete })}
        >
            {metadataKey}
        </Badge>
        <dd className={classNames('d-flex m-0', { ['d-none']: !metadataValue || !onDelete })}>
            {metadataValue && (
                <Badge
                    small={small}
                    variant="secondary"
                    aria-label={`Repository metadata value "${metadataValue}" for key=${metadataKey}`}
                    className={classNames(styles.roundedLeftNone, 'm-0', {
                        [styles.roundedRightNone]: onDelete,
                    })}
                >
                    {metadataValue}
                </Badge>
            )}
            {onDelete && (
                <Button
                    size="sm"
                    className={classNames('px-1 py-0 m-0', styles.roundedLeftNone)}
                    variant="warning"
                    outline={true}
                    aria-label="Delete repository metadata"
                    onClick={onDelete}
                >
                    <Icon svgPath={mdiClose} aria-hidden={true} />
                </Button>
            )}
        </dd>
    </div>
)

interface RepoMetadataProps {
    metadata: [string, string | null | undefined][]
    className?: string
    small?: boolean
    onDelete?: (key: string, value: string | null | undefined) => void
}

export const RepoMetadata: React.FunctionComponent<RepoMetadataProps> = ({ metadata, small, className, onDelete }) => {
    const sortedPairs = useMemo(() => sortBy(metadata), [metadata])
    if (sortedPairs.length === 0) {
        return null
    }
    return (
        <dl className={classNames(styles.repoMetadata, 'd-flex align-items-start flex-wrap m-0', className)}>
            {sortedPairs.map(([key, value]) => (
                <RepoMetadataItem
                    key={`${key}:${value}`}
                    metadataKey={key}
                    metadataValue={value}
                    small={small}
                    onDelete={onDelete ? () => onDelete?.(key, value) : undefined}
                />
            ))}
        </dl>
    )
}
