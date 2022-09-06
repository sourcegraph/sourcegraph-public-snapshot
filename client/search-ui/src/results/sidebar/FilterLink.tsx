import React from 'react'

import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { SyntaxHighlightedSearchQuery, CodeHostIcon } from '@sourcegraph/search-ui'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import { getFiltersOfKind } from './helpers'

import styles from './SearchFilterSection.module.scss'

export interface FilterLinkProps {
    label: string
    ariaLabel?: string
    value: string
    count?: number
    limitHit?: boolean
    kind?: string
    labelConverter?: (label: string) => JSX.Element
    onFilterChosen: (value: string, kind?: string) => void
}

export const FilterLink: React.FunctionComponent<React.PropsWithChildren<FilterLinkProps>> = ({
    label,
    ariaLabel,
    value,
    count,
    limitHit,
    kind,
    labelConverter = label => (label === value ? <SyntaxHighlightedSearchQuery query={label} /> : label),
    onFilterChosen,
}) => {
    const countTooltip = count
        ? `At least ${count} ${pluralize('result matches', count, 'results match')} this filter.`
        : ''

    return (
        <Button
            className={styles.sidebarSectionListItem}
            onClick={() => onFilterChosen(value, kind)}
            data-testid="filter-link"
            value={value}
            variant="link"
            aria-label={`${ariaLabel ?? label}${countTooltip ? `, ${countTooltip}` : ''}`}
        >
            <span className={classNames('flex-grow-1', styles.sidebarSectionListItemLabel)}>
                {labelConverter(label)}
            </span>
            {count && (
                <Tooltip content={countTooltip}>
                    <span className="pl-2 flex-shrink-0">
                        {count}
                        {limitHit ? '+' : ''}
                    </span>
                </Tooltip>
            )}
        </Button>
    )
}

export const getRepoFilterLinks = (
    filters: Filter[] | undefined,
    onFilterChosen: (value: string, kind?: string) => void,
    coreWorkflowImprovementsEnabled: boolean | undefined
): React.ReactElement[] => {
    function repoLabelConverter(label: string): JSX.Element {
        const Icon = CodeHostIcon({
            repoName: label,
            className: classNames(!coreWorkflowImprovementsEnabled && 'text-muted', styles.sidebarSectionIcon),
        })

        return (
            <span
                className={classNames(
                    !coreWorkflowImprovementsEnabled && 'text-monospace search-query-link',
                    styles.sidebarSectionListItemBreakWords
                )}
            >
                {!coreWorkflowImprovementsEnabled && <span className="search-filter-keyword">r:</span>}
                {Icon ? (
                    <>
                        {Icon}
                        {displayRepoName(label)}
                    </>
                ) : (
                    label
                )}
            </span>
        )
    }

    return getFiltersOfKind(filters, FilterType.repo).map(filter => (
        <FilterLink
            {...filter}
            key={`${filter.label}-${filter.value}`}
            labelConverter={repoLabelConverter}
            onFilterChosen={onFilterChosen}
            ariaLabel={`Search in repository ${filter.label}`}
        />
    ))
}

export const getDynamicFilterLinks = (
    filters: Filter[] | undefined,
    kinds: Filter['kind'][],
    onFilterChosen: (value: string, kind?: string) => void,
    ariaLabelTransform: (label: string, value: string) => string = label => `${label}`,
    labelTransform: (label: string, value: string) => string = label => `${label}`
): React.ReactElement[] =>
    (filters || [])
        .filter(filter => kinds.includes(filter.kind))
        .map(filter => (
            <FilterLink
                {...filter}
                label={labelTransform(filter.label, filter.value)}
                ariaLabel={ariaLabelTransform(filter.label, filter.value)}
                key={`${filter.label}-${filter.value}`}
                onFilterChosen={onFilterChosen}
            />
        ))

export const getSearchSnippetLinks = (
    settingsCascade: SettingsCascadeProps['settingsCascade'],
    onFilterChosen: (value: string) => void
): React.ReactElement[] => {
    const snippets = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final['search.scopes']) || []
    return snippets.map(snippet => (
        <FilterLink
            label={snippet.name}
            value={snippet.value}
            key={`${snippet.name}-${snippet.value}`}
            onFilterChosen={onFilterChosen}
            ariaLabel={`Use search snippet: ${snippet.name}`}
        />
    ))
}
