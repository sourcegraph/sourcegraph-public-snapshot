import * as React from 'react'
import { RepositoryIcon } from '../../../../shared/src/components/icons'
import * as GQL from '../../../../shared/src/graphql/schema'
import { LanguageIcon } from '../../../../shared/src/components/languageIcons'
import { dirname, basename } from '../../util/path'
import FilterIcon from 'mdi-react/FilterIcon'
import FileIcon from 'mdi-react/FileIcon'
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
export const fuzzySearchFilters = [SuggestionTypes.repo, SuggestionTypes.repogroup]

/**
 * dir and symbol are fetched/suggested by the fuzzy-search
 * but are not filter types: /web/src/search/searchFilterSuggestions.ts
 */
export type FiltersSuggestionTypes = Exclude<SuggestionTypes, SuggestionTypes.dir | SuggestionTypes.symbol>

export interface Suggestion {
    title: string
    description?: string
    type: SuggestionTypes
    url?: string
    label?: string
    kind?: GQL.SymbolKind
}

interface SuggestionIconProps {
    suggestion: Suggestion
    className?: string
    size: number
}

export function createSuggestion(item: GQL.SearchSuggestion): Suggestion | undefined {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: SuggestionTypes.repo,
                title: item.name,
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
                    title: item.name,
                    description: descriptionParts.join(' — '),
                    url: `${item.url}?suggestion`,
                    label: 'go to dir',
                }
            }
            return {
                type: SuggestionTypes.file,
                title: item.name,
                description: descriptionParts.join(' — '),
                url: `${item.url}?suggestion`,
                label: 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: SuggestionTypes.symbol,
                kind: item.kind,
                title: item.name,
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

const SuggestionIcon: React.FunctionComponent<SuggestionIconProps> = ({ suggestion, children, ...passThru }) => {
    switch (suggestion.type) {
        case SuggestionTypes.filters:
            return <FilterIcon {...passThru} />
        case SuggestionTypes.repo:
            return <RepositoryIcon {...passThru} />
        case SuggestionTypes.file:
            return <FileIcon {...passThru} />
        case SuggestionTypes.lang:
            return <LanguageIcon {...passThru} language={suggestion.title} {...passThru} />
        case SuggestionTypes.symbol:
            if (!suggestion.kind) {
                return null
            }
            return <SymbolIcon kind={suggestion.kind} {...passThru} />
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
        <SuggestionIcon size={20} className="icon-inline suggestion__icon" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.title}</div>
        <div className="suggestion__description">{suggestion.description}</div>
        {showUrlLabel && !!suggestion.label && (
            <div className="suggestion__action" hidden={!isSelected}>
                <kbd>enter</kbd> {suggestion.label}
            </div>
        )}
    </li>
)
