import * as React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { LanguageIcon } from '../../../../shared/src/components/languageIcons'
import { dirname, basename } from '../../util/path'
import FilterIcon from 'mdi-react/FilterIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'

export enum SuggestionTypes {
    filters = 'filters',
    repo = 'repo',
    repogroup = 'repogroup',
    repohasfile = 'repohasfile',
    repohascommitafter = 'repohascommitafter',
    file = 'file',
    type = 'type',
    case = 'case',
    lang = 'lang',
    fork = 'fork',
    archived = 'archived',
    count = 'count',
    timeout = 'timeout',
    dir = 'dir',
    symbol = 'symbol',
}

/**
 * Filters which use fuzzy-search for their suggestion values
 */
export const fuzzySearchFilters = [
    SuggestionTypes.repo,
    SuggestionTypes.repogroup,
    SuggestionTypes.file,
    SuggestionTypes.repohasfile,
]

/**
 * dir and symbol are fetched/suggested by the fuzzy-search
 * but are not filter types: /web/src/search/searchFilterSuggestions.ts
 */
export type FiltersSuggestionTypes = Exclude<SuggestionTypes, SuggestionTypes.dir | SuggestionTypes.symbol>

export interface Suggestion {
    type: SuggestionTypes
    /** The value to be suggested and that will be added to queries */
    value: string
    /** Description that will be displayed together with suggestion value */
    description?: string
    /** Fuzzy-search suggestions may have a url for redirect when selected */
    url?: string
    /** Label informing what will happen when suggestion is selected */
    label?: string
    /** For suggestions of type `symbol` */
    symbolKind?: GQL.SymbolKind
}

interface SuggestionIconProps {
    suggestion: Suggestion
    className?: string
}

export function createSuggestion(item: GQL.SearchSuggestion): Suggestion | undefined {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: SuggestionTypes.repo,
                value: item.name,
                url: `/${item.name}`,
                label: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts = []
            const dir = dirname(item.path)
            if (dir !== '.') {
                descriptionParts.push(`${dir}/`)
            }
            descriptionParts.push(basename(item.repository.name))
            if (item.isDirectory) {
                return {
                    type: SuggestionTypes.dir,
                    value: item.name,
                    description: descriptionParts.join(' — '),
                    url: `${item.url}?suggestion`,
                    label: 'go to dir',
                }
            }
            return {
                type: SuggestionTypes.file,
                value: item.name,
                description: descriptionParts.join(' — '),
                url: `${item.url}?suggestion`,
                label: 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: SuggestionTypes.symbol,
                symbolKind: item.kind,
                value: item.name,
                description: `${item.containerName || item.location.resource.path} — ${basename(
                    item.location.resource.repository.name
                )}`,
                url: item.url,
                label: 'go to definition',
            }
        }
        default:
            return undefined
    }
}

const SuggestionIcon: React.FunctionComponent<SuggestionIconProps> = ({ suggestion, children, ...props }) => {
    switch (suggestion.type) {
        case SuggestionTypes.filters:
            return <FilterIcon {...props} />
        case SuggestionTypes.repo:
            return <SourceRepositoryIcon {...props} />
        case SuggestionTypes.file:
            return <FileIcon {...props} />
        case SuggestionTypes.lang:
            return <LanguageIcon {...props} language={suggestion.value} {...props} />
        case SuggestionTypes.symbol:
            if (!suggestion.symbolKind) {
                return null
            }
            return <SymbolIcon kind={suggestion.symbolKind} {...props} />
        default:
            return null
    }
}

interface SuggestionProps {
    suggestion: Suggestion
    isSelected?: boolean
    onClick?: () => void
    showUrlLabel: boolean
}

export const SuggestionItem: React.FunctionComponent<SuggestionProps> = ({
    suggestion,
    isSelected,
    showUrlLabel,
    ...props
}) => (
    <li className={'suggestion' + (isSelected ? ' suggestion--selected' : '')} {...props}>
        <SuggestionIcon className="icon-inline suggestion__icon" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.value}</div>
        <div className="suggestion__description">{suggestion.description}</div>
        {showUrlLabel && !!suggestion.label && (
            <div className="suggestion__action" hidden={!isSelected}>
                <kbd>enter</kbd> {suggestion.label}
            </div>
        )}
    </li>
)
