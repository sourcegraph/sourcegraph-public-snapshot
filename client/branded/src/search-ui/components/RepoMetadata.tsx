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

export function GetFilterLink(
    pred: string,
    key: string,
    value?: string | null,
    queryState?: QueryState,
    omitRepoFilter?: boolean,
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
): string | undefined {
    if (!queryState || !buildSearchURLQueryFromQueryState) {
        return undefined
    }

    let query = queryState.query
    // omit repo: filter if omitRepoFilter is true
    if (omitRepoFilter) {
        const repoFilter = findFilter(queryState.query, 'repo', FilterKind.Global)
        if (repoFilter && !repoFilter.value?.value.startsWith(pred)) {
            query = omitFilter(query, repoFilter)
        }
    }
    // append metadata filter
    query = appendFilter(query, 'repo', value ? `${pred}(${key}:${value})` : `${pred}(${key})`)

    const searchParams = buildSearchURLQueryFromQueryState({ query })
    return `/search?${searchParams}`
}

const Meta: React.FC<MetaProps> = ({
    meta,
    queryState,
    buildSearchURLQueryFromQueryState,
    onDelete,
    queryBuildOptions,
}) => {
    const filterLink = useMemo(
        () =>
            GetFilterLink(
                'has.meta',
                meta.key,
                meta.value,
                queryState,
                queryBuildOptions?.omitRepoFilter,
                buildSearchURLQueryFromQueryState
            ),
        [buildSearchURLQueryFromQueryState, meta.key, meta.value, queryBuildOptions?.omitRepoFilter, queryState]
    )

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

export interface Tag {
    key: string
    value?: string | null
    filterLink?: string
}

interface TagListProps {
    tags: Tag[]
    className?: string
}

const TagContent: React.FC<{ tag: Tag; highlight?: boolean }> = ({ tag, highlight }) => (
    <Code>
        <span aria-label="Repository metadata key" className={classNames({ [styles.highlight]: highlight })}>
            {tag.key}
        </span>
        {tag.value ? (
            <span aria-label="Repository metadata value">:{tag.value}</span>
        ) : (
            <VisuallyHidden>No metadata value</VisuallyHidden>
        )}
    </Code>
)

// TagList is used for displaying tags (metadata or topics) for repository search results and on the repository's //
// tree page.
export const TagList: React.FC<TagListProps> = ({ tags, className }) => {
    const sortedTags = useMemo(() => sortBy(tags, ['key', 'value']), [tags])
    if (sortedTags.length === 0) {
        return null
    }
    return (
        <ul className={classNames(styles.container, 'd-flex align-items-start flex-wrap m-0 list-unstyled', className)}>
            {sortedTags.map(tag => (
                <li key={`${tag.key}:${tag.value}`}>
                    {tag.filterLink ? (
                        <Badge variant="outlineSecondary" as={Link} to={tag.filterLink}>
                            <TagContent tag={tag} highlight={true} />
                        </Badge>
                    ) : (
                        <Badge variant="outlineSecondary">
                            <TagContent tag={tag} />
                        </Badge>
                    )}
                </li>
            ))}
        </ul>
    )
}

export function topicToTag(
    topic: string,
    queryState?: QueryState,
    omitRepoFilter?: boolean,
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
): Tag {
    return {
        key: topic,
        value: undefined,
        filterLink: GetFilterLink(
            'has.topic',
            topic,
            undefined,
            queryState,
            omitRepoFilter,
            buildSearchURLQueryFromQueryState
        ),
    }
}

export function metadataToTag(
    meta: RepoMetadataItem,
    queryState?: QueryState,
    omitRepoFilter?: boolean,
    buildSearchURLQueryFromQueryState?: (queryParameters: BuildSearchQueryURLParameters) => string
): Tag {
    return {
        key: meta.key,
        value: meta.value,
        filterLink: GetFilterLink(
            'has.meta',
            meta.key,
            meta.value,
            queryState,
            omitRepoFilter,
            buildSearchURLQueryFromQueryState
        ),
    }
}
