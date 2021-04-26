import classNames from 'classnames'
import React from 'react'

import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { SyntaxHighlightedSearchQuery } from '../../../../components/SyntaxHighlightedSearchQuery'
import { Settings } from '../../../../schema/settings.schema'
import { AggregateStreamingSearchResults } from '../../../stream'

import styles from './SearchSidebarSection.module.scss'

interface FilterLinkProps {
    label: string
    value: string
    count?: number
    limitHit?: boolean
    onFilterChosen: (value: string) => void
}

const FilterLink: React.FunctionComponent<FilterLinkProps> = ({ label, value, count, limitHit, onFilterChosen }) => (
    <button
        type="button"
        className={classNames('btn btn-link', styles.sidebarSectionListItem)}
        data-tooltip={value}
        data-placement="right"
        onClick={() => onFilterChosen(value)}
    >
        <span className="flex-grow-1">{label === value ? <SyntaxHighlightedSearchQuery query={label} /> : label}</span>
        <span className="pl-1 flex-shrink-0">
            {count}
            {limitHit ? '+' : ''}
        </span>
    </button>
)

export const getRepoFilterLinks = (
    results: AggregateStreamingSearchResults | undefined,
    onFilterChosen: (value: string) => void
): React.ReactElement[] => {
    if (results?.filters) {
        return results?.filters
            .filter(filter => filter.kind === 'repo' && filter.value !== '')
            .map(filter => (
                <FilterLink {...filter} key={`${filter.label}-${filter.value}`} onFilterChosen={onFilterChosen} />
            ))
    }
    return []
}

export const getDynamicFilterLinks = (
    results: AggregateStreamingSearchResults | undefined,
    onFilterChosen: (value: string) => void
): React.ReactElement[] => {
    if (results?.filters) {
        return results?.filters
            .filter(filter => filter.kind !== 'repo')
            .map(filter => (
                <FilterLink {...filter} key={`${filter.label}-${filter.value}`} onFilterChosen={onFilterChosen} />
            ))
    }
    return []
}

export const getSnippets = (
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
