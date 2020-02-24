import * as React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { LanguageIcon } from '../../../../shared/src/components/languageIcons'
import { dirname, basename } from '../../util/path'
import FilterIcon from 'mdi-react/FilterIcon'
import FileIcon from 'mdi-react/FileIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { SuggestionTypes, NonFilterSuggestionTypes } from '../../../../shared/src/search/suggestions/util'
import { escapeRegExp } from 'lodash'
import { FilterTypes } from '../../../../shared/src/search/interactive/util'

export const filterAliases: Record<string, FiltersSuggestionTypes | undefined> = {
    r: FilterTypes.repo,
    g: FilterTypes.repogroup,
    f: FilterTypes.file,
    l: FilterTypes.lang,
    language: FilterTypes.lang,
}

/**
 * Filters which use fuzzy-search for their suggestion values
 */
export const fuzzySearchFilters: FiltersSuggestionTypes[] = [
    FilterTypes.repo,
    FilterTypes.repogroup,
    FilterTypes.file,
    FilterTypes.repohasfile,
]

/**
 * Some filter types should have their suggestions searched without influence
 * from the rest of the query, as they will then influence the scope of other filters.
 */
export const isolatedFuzzySearchFilters: FiltersSuggestionTypes[] = [FilterTypes.repo, FilterTypes.repogroup]

/**
 * dir and symbol are fetched/suggested by the fuzzy-search
 * but are not filter types: /web/src/search/searchFilterSuggestions.ts
 */
export type FiltersSuggestionTypes = FilterTypes | 'filters'

export interface Suggestion {
    type: SuggestionTypes
    /** The value to be suggested and that will be added to queries */
    value: string
    /**
     * Optional value to use when suggestion is displayed to the user.
     * Useful for displaying a "human-readable" suggestion value
     */
    displayValue?: string
    /** Description that will be displayed together with suggestion value */
    description?: string
    /** Fuzzy-search suggestions may have a url for redirect when selected */
    url?: string
    /** Label informing what will happen when suggestion is selected */
    label?: string
    /** For suggestions of type `symbol` */
    symbolKind?: GQL.SymbolKind
    /** If the suggestion was loaded from the fuzzy-search */
    fromFuzzySearch?: true
}

interface SuggestionIconProps {
    suggestion: Suggestion
    className?: string
}

/**
 * @returns The given string with escaped special characters and wrapped with regex boundaries
 */
const formatRegExp = (value: string): string => '^' + escapeRegExp(value) + '$'

export function createSuggestion(item: GQL.SearchSuggestion): Suggestion | undefined {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: FilterTypes.repo,
                // Add "regex start and end boundaries" to
                // correctly scope additional suggestions
                value: formatRegExp(item.name),
                displayValue: item.name,
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
                    type: NonFilterSuggestionTypes.dir,
                    value: '^' + escapeRegExp(item.path),
                    description: descriptionParts.join(' — '),
                    url: `${item.url}?suggestion`,
                    label: 'go to dir',
                }
            }
            return {
                type: FilterTypes.file,
                value: formatRegExp(item.path),
                displayValue: item.name,
                description: descriptionParts.join(' — '),
                url: `${item.url}?suggestion`,
                label: 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: NonFilterSuggestionTypes.symbol,
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
        case NonFilterSuggestionTypes.filters:
            return <FilterIcon {...props} />
        case FilterTypes.repo:
            return <SourceRepositoryIcon {...props} />
        case FilterTypes.file:
            return <FileIcon {...props} />
        case FilterTypes.lang:
            return <LanguageIcon {...props} language={suggestion.value} {...props} />
        case NonFilterSuggestionTypes.symbol:
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
    /** If suggestion.label should be shown, else show defaultLabel  */
    showUrlLabel: boolean
    /** Suggestion label to show if(!showUrlLabel || !suggestion.label) */
    defaultLabel?: string
}

export const SuggestionItem: React.FunctionComponent<SuggestionProps> = ({
    suggestion,
    isSelected,
    showUrlLabel,
    defaultLabel,
    ...props
}) => (
    <li className={`suggestion ${isSelected ? ' suggestion--selected' : ''} e2e-suggestion-item`} {...props}>
        <SuggestionIcon className="icon-inline suggestion__icon" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.displayValue ?? suggestion.value}</div>
        <div className="suggestion__description">{suggestion.description}</div>
        {(showUrlLabel || defaultLabel) && (
            <div className="suggestion__action" hidden={!isSelected}>
                <kbd>enter</kbd> {(showUrlLabel && suggestion?.label) || defaultLabel}
            </div>
        )}
    </li>
)
