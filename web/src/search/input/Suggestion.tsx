import FolderIcon from 'mdi-react/FolderIcon'
import * as React from 'react'
import { RepositoryIcon } from '../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import * as GQL from '../../../../shared/src/graphql/schema'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
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

interface SymbolSuggestion extends BaseSuggestion {
    type: 'symbol'
    kind: GQL.SymbolKind
}

interface RepoSuggestion extends BaseSuggestion {
    type: 'repo'
}

interface FileSuggestion extends BaseSuggestion {
    type: 'file'
}

interface DirSuggestion extends BaseSuggestion {
    type: 'dir'
}

export type Suggestion = SymbolSuggestion | RepoSuggestion | FileSuggestion | DirSuggestion

export function createSuggestion(lineNumber: number, item: GQL.SearchSuggestion): Suggestion {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: 'repo',
                title: item.name,
                url: `/${item.name}`,
                urlLabel: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts = []
            const dir = dirname(item.path)
            if (dir !== undefined && dir !== '.') {
                descriptionParts.push(`${dir}/`)
            }
            descriptionParts.push(basename(item.repository.name))
            if (item.isDirectory) {
                return {
                    type: 'dir',
                    title: item.name,
                    description: descriptionParts.join(' — '),
                    url: `${item.url}?suggestion`,
                    urlLabel: 'go to dir',
                }
            }
            const withLineNumber = lineNumber > 0
            return {
                type: 'file',
                title: withLineNumber ? `${item.name} at line ${lineNumber}` : item.name,
                description: descriptionParts.join(' — '),
                url: withLineNumber ? `${item.url}#L${lineNumber}` : `${item.url}?suggestion`,
                urlLabel: withLineNumber ? `go to line ${lineNumber}` : 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: 'symbol',
                kind: item.kind,
                title: item.name,
                description: `${item.containerName || item.location.resource.path} — ${basename(
                    item.location.resource.repository.name
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

const SuggestionIcon: React.FunctionComponent<SuggestionIconProps> = ({ suggestion, ...passThru }) => {
    switch (suggestion.type) {
        case 'repo':
            return <RepositoryIcon {...passThru} />
        case 'dir':
            return <FolderIcon />
        case 'file':
            return <SymbolIcon kind={GQL.SymbolKind.FILE} {...passThru} />
        case 'symbol':
            return <SymbolIcon kind={suggestion.kind} {...passThru} />
    }
}

interface SuggestionProps {
    suggestion: Suggestion

    isSelected?: boolean

    /** Called when the user clicks on the suggestion */
    onClick?: () => void

    /** Get a reference to the HTML element for scroll management */
    liRef?: (ref: HTMLLIElement | null) => void
}

export const SuggestionItem = ({ suggestion, isSelected, onClick, liRef }: SuggestionProps) => (
    <li className={'suggestion' + (isSelected ? ' suggestion--selected' : '')} onMouseDown={onClick} ref={liRef}>
        <SuggestionIcon className="icon-inline suggestion__icon" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.title}</div>
        <div className="suggestion__description">{suggestion.description}</div>
        <div className="suggestion__action" hidden={!isSelected}>
            <kbd>enter</kbd> {suggestion.urlLabel}
        </div>
    </li>
)
