import React, { useMemo } from 'react'

import { mdiDelete } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { sortBy } from 'lodash'

import type { BuildSearchQueryURLParameters, QueryState } from '@sourcegraph/shared/src/search'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendFilter, omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { Badge, Button, Code, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

export interface RepoMetadataItem {
    key: string
    value?: string | null
}

const MetaContent: React.FC<{ meta: RepoMetadataItem; highlight?: boolean }> = ({ meta, highlight }) => (
    <Code>
        <span aria-label="Repository metadata key" className={classNames({ [styles.highlight]: highlight })}>
            {meta.key}
        </span>
        {meta.value ? (
            <span aria-label="Repository metadata value">:{meta.value}</span>
        ) : (
            <VisuallyHidden>No metadata value</VisuallyHidden>
        )}
    </Code>
)

interface MetaProps {
    meta: RepoMetadataItem
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
    queryState?: QueryState
    queryBuildOptions?: {
        omitRepoFilter?: boolean
    }
    onDelete?: (item: RepoMetadataItem) => void
}

const Meta: React.FC<MetaProps> = ({
    meta,
    queryState,
    buildSearchURLQueryFromQueryState,
    onDelete,
    queryBuildOptions,
}) => {
    const filterLink = useMemo(() => {
        if (!queryState || !buildSearchURLQueryFromQueryState) {
            return undefined
        }

        let query = queryState.query
        // omit repo: filter if omitRepoFilter is true
        if (queryBuildOptions?.omitRepoFilter) {
            const repoFilter = findFilter(queryState.query, 'repo', FilterKind.Global)
            if (repoFilter && !repoFilter.value?.value.startsWith('has.meta')) {
                query = omitFilter(query, repoFilter)
            }
        }
        // append metadata filter
        query = appendFilter(
            query,
            'repo',
            meta.value ? `has.meta(${meta.key}:${meta.value})` : `has.meta(${meta.key})`
        )

        const searchParams = buildSearchURLQueryFromQueryState({ query })
        return `/search?${searchParams}`
    }, [buildSearchURLQueryFromQueryState, meta.key, meta.value, queryBuildOptions?.omitRepoFilter, queryState])

    if (onDelete) {
        return (
            <Tooltip content="Remove from this repository">
                <Badge
                    variant="outlineSecondary"
                    as={Button}
                    onClick={() => onDelete(meta)}
                    aria-label="Remove from this repository"
                    className={styles.badgeButton}
                >
                    <Icon svgPath={mdiDelete} aria-hidden={true} className="mr-1" />
                    <MetaContent meta={meta} />
                </Badge>
            </Tooltip>
        )
    }

    if (filterLink) {
        return (
            <Badge variant="outlineSecondary" as={Link} to={filterLink}>
                <MetaContent meta={meta} highlight={true} />
            </Badge>
        )
    }

    return (
        <Badge variant="outlineSecondary">
            <MetaContent meta={meta} />
        </Badge>
    )
}

interface RepoMetadataProps
    extends Pick<MetaProps, 'buildSearchURLQueryFromQueryState' | 'queryState' | 'onDelete' | 'queryBuildOptions'> {
    items: RepoMetadataItem[]
    className?: string
}

export const RepoMetadata: React.FC<RepoMetadataProps> = ({ items, className, onDelete, ...props }) => {
    const sortedItems = useMemo(() => sortBy(items, ['key', 'value']), [items])
    if (sortedItems.length === 0) {
        return null
    }

    return (
        <ul className={classNames(styles.container, 'd-flex align-items-start flex-wrap m-0 list-unstyled', className)}>
            {sortedItems.map(item => (
                <li key={`${item.key}:${item.value}`}>
                    <Meta meta={item} onDelete={onDelete} {...props} />
                </li>
            ))}
        </ul>
    )
}
