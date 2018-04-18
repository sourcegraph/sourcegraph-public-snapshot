import FolderIcon from '@sourcegraph/icons/lib/Folder'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { SymbolIcon } from '../../symbols/SymbolIcon'
import { basename, dirname } from '../../util/path'

interface BaseSuggestion {
    title: string
    description?: string

    /**
     * The URL that is navigated to when the user selects this
     * suggestion.
     */
    url: string

    /**
     * A label describing the action taken when navigating to
     * the URL (e.g., "go to repository").
     */
    urlLabel: string
}

export interface SymbolSuggestion extends BaseSuggestion {
    type: 'symbol'
    kind: GQL.SymbolKind
}

export interface RepoSuggestion extends BaseSuggestion {
    type: 'repo'
}

export interface FileSuggestion extends BaseSuggestion {
    type: 'file'
}

export interface DirSuggestion extends BaseSuggestion {
    type: 'dir'
}

export type Suggestion = SymbolSuggestion | RepoSuggestion | FileSuggestion | DirSuggestion

export function createSuggestion(item: GQL.SearchSuggestion): Suggestion {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: 'repo',
                title: item.uri,
                url: `/${item.uri}`,
                urlLabel: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts = []
            const dir = dirname(item.path)
            if (dir !== undefined && dir !== '.') {
                descriptionParts.push(`${dir}/`)
            }
            descriptionParts.push(basename(item.repository.uri))
            if (item.isDirectory) {
                return {
                    type: 'dir',
                    title: item.name,
                    description: descriptionParts.join(' — '),
                    url: item.url,
                    urlLabel: 'go to dir',
                }
            }
            return {
                type: 'file',
                title: item.name,
                description: descriptionParts.join(' — '),
                url: item.url,
                urlLabel: 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: 'symbol',
                kind: item.kind,
                title: item.name,
                description: `${item.containerName || item.location.resource.path} — ${basename(
                    item.location.resource.repository.uri
                )}`,
                url: item.url,
                urlLabel: 'go to definition',
            }
        }
    }
}

interface SuggestionIconProps {
    suggestion: Suggestion
    className?: string
}

const SuggestionIcon: React.StatelessComponent<SuggestionIconProps> = ({ suggestion, ...passThru }) => {
    switch (suggestion.type) {
        case 'repo':
            return <RepoIcon {...passThru} />
        case 'dir':
            return <FolderIcon {...passThru} />
        case 'file':
            return <SymbolIcon kind={GQL.SymbolKind.FILE} {...passThru} />
        case 'symbol':
            return <SymbolIcon kind={suggestion.kind} {...passThru} />
    }
}

export interface SuggestionProps {
    suggestion: Suggestion

    isSelected?: boolean

    /** Called when the user clicks on the suggestion */
    onClick?: () => void

    /** Get a reference to the HTML element for scroll management */
    liRef?: (ref: HTMLLIElement | null) => void
}

export const SuggestionItem = ({ suggestion, isSelected, onClick, liRef }: SuggestionProps) => (
    <li className={'suggestion' + (isSelected ? ' suggestion--selected' : '')} onMouseDown={onClick} ref={liRef}>
        <SuggestionIcon className="icon-inline" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.title}</div>
        <div className="suggestion__description">{suggestion.description}</div>
        <div className="suggestion__action" hidden={!isSelected}>
            <kbd>enter</kbd> {suggestion.urlLabel}
        </div>
    </li>
)
