import React, { useMemo } from 'react'

import { mdiDelete } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { sortBy } from 'lodash'

import { BuildSearchQueryURLParameters, QueryState } from '@sourcegraph/shared/src/search'
import { appendFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { Badge, Button, Code, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

export interface RepoMetadataItem {
    key: string
    value?: string | null
}

interface MetaProps {
    meta: RepoMetadataItem
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
    queryState?: QueryState
}

const Meta: React.FC<MetaProps> = ({ meta: { key, value }, queryState, buildSearchURLQueryFromQueryState }) => {
    const filterLink = useMemo(() => {
        if (!queryState || !buildSearchURLQueryFromQueryState) {
            return undefined
        }

        const query = appendFilter(queryState.query, 'repo', `has.key(${key})`)

        const searchParams = buildSearchURLQueryFromQueryState({ query })
        return `/search?${searchParams}`
    }, [buildSearchURLQueryFromQueryState, key, queryState])

    return (
        <Code className={styles.meta}>
            {filterLink ? (
                <Link aria-label="Filter by repository metadata" className={styles.highlight} to={filterLink}>
                    {key}
                </Link>
            ) : (
                <span aria-label="Repository metadata key">{key}</span>
            )}
            {value ? (
                <span aria-label="Repository metadata value">:{value}</span>
            ) : (
                <VisuallyHidden>No metadata value</VisuallyHidden>
            )}
        </Code>
    )
}

interface RepoMetadataProps extends Pick<MetaProps, 'buildSearchURLQueryFromQueryState' | 'queryState'> {
    items: RepoMetadataItem[]
    className?: string
    onDelete?: (item: RepoMetadataItem) => void
    small?: boolean
}

export const RepoMetadata: React.FC<RepoMetadataProps> = ({
    items,
    className,
    onDelete,
    small,
    queryState,
    buildSearchURLQueryFromQueryState,
}) => {
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
                                <Meta meta={item} />
                            </Button>
                        </Tooltip>
                    ) : (
                        <Badge key={`${item.key}:${item.value}`} className="d-flex">
                            <Meta
                                meta={item}
                                queryState={queryState}
                                buildSearchURLQueryFromQueryState={buildSearchURLQueryFromQueryState}
                            />
                        </Badge>
                    )}
                </li>
            ))}
        </ul>
    )
}
