import classNames from 'classnames'
import React from 'react'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { SyntaxHighlightedSearchQuery } from '../../../components/SyntaxHighlightedSearchQuery'
import { Settings } from '../../../schema/settings.schema'

import styles from './SearchSidebarSection.module.scss'

export interface FilterLinkProps {
    label: string
    value: string
    count?: number
    limitHit?: boolean
    labelConverter?: (label: string) => JSX.Element
    onFilterChosen: (value: string) => void
}

export const FilterLink: React.FunctionComponent<FilterLinkProps> = ({
    label,
    value,
    count,
    limitHit,
    labelConverter = label => (label === value ? <SyntaxHighlightedSearchQuery query={label} /> : label),
    onFilterChosen,
}) => (
    <button
        type="button"
        className={classNames('btn btn-link', styles.sidebarSectionListItem)}
        onClick={() => onFilterChosen(value)}
        data-testid="filter-link"
        value={value}
    >
        <span className="flex-grow-1">{labelConverter(label)}</span>
        {count && (
            <span
                className="pl-2 flex-shrink-0"
                title={`At least ${count} ${pluralize('result matches', count, 'results match')} this filter.`}
            >
                {count}
                {limitHit ? '+' : ''}
            </span>
        )}
    </button>
)

export const getRepoFilterLinks = (
    filters: Filter[] | undefined,
    onFilterChosen: (value: string) => void
): React.ReactElement[] => {
    function repoLabelConverter(label: string): JSX.Element {
        const Icon = RepoIcon({
            repoName: label,
            className: classNames('icon-inline text-muted', styles.sidebarSectionIcon),
        })

        return (
            <span className={classNames('text-monospace search-query-link', styles.sidebarSectionListItemBreakWords)}>
                <span className="search-filter-keyword">r:</span>
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

    return (filters || [])
        .filter(filter => filter.kind === 'repo' && filter.value !== '')
        .map(filter => (
            <FilterLink
                {...filter}
                key={`${filter.label}-${filter.value}`}
                labelConverter={repoLabelConverter}
                onFilterChosen={onFilterChosen}
            />
        ))
}

export const getDynamicFilterLinks = (
    filters: Filter[] | undefined,
    onFilterChosen: (value: string) => void
): React.ReactElement[] =>
    (filters || [])
        .filter(filter => filter.kind !== 'repo')
        .map(filter => (
            <FilterLink {...filter} key={`${filter.label}-${filter.value}`} onFilterChosen={onFilterChosen} />
        ))

export const getSearchScopeLinks = (
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
        />
    ))
}
