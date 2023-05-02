import React, { useMemo } from 'react'

import { mdiDelete } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { sortBy } from 'lodash'

import { Badge, Button, Code, Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

export interface RepoMetadataItem {
    key: string
    value?: string | null
}

const Meta: React.FC<React.PropsWithChildren<{ meta: RepoMetadataItem; highlight?: boolean }>> = ({
    meta: { key, value },
    highlight = true,
}) => (
    <Code className={styles.meta}>
        <span aria-label="Repository metadata key" className={classNames({ [styles.highlight]: highlight })}>
            {key}
        </span>
        {value ? (
            <span aria-label="Repository metadata value">:{value}</span>
        ) : (
            <VisuallyHidden>No metadata value</VisuallyHidden>
        )}
    </Code>
)

interface RepoMetadataProps {
    items: RepoMetadataItem[]
    className?: string
    onDelete?: (item: RepoMetadataItem) => void
    small?: boolean
}

export const RepoMetadata: React.FC<RepoMetadataProps> = ({ items, className, onDelete, small }) => {
    const sortedItems = useMemo(() => sortBy(items, ['key', 'value']), [items])
    if (sortedItems.length === 0) {
        return null
    }

    return (
        <ul
            className={classNames(styles.container, 'd-flex align-items-start flex-wrap m-0 list-unstyled', className, {
                [styles.small]: small,
            })}
        >
            {sortedItems.map(item => (
                <li key={`${item.key}:${item.value}`}>
                    {onDelete ? (
                        <Tooltip content="Delete metadata">
                            <Button
                                className={styles.deleteButton}
                                variant="secondary"
                                outline={true}
                                size="sm"
                                aria-label="Delete metadata"
                                onClick={() => onDelete(item)}
                            >
                                <Icon svgPath={mdiDelete} aria-hidden={true} className="mr-1" />
                                <Meta meta={item} highlight={false} />
                            </Button>
                        </Tooltip>
                    ) : (
                        <Badge key={`${item.key}:${item.value}`} className="d-flex">
                            <Meta meta={item} />
                        </Badge>
                    )}
                </li>
            ))}
        </ul>
    )
}
