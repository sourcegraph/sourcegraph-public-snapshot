import BookIcon from '@sourcegraph/icons/lib/Book'
import FileIcon from '@sourcegraph/icons/lib/File'
import FolderIcon from '@sourcegraph/icons/lib/Folder'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { basename, dirname } from '../util/path'

export const enum SuggestionType {
    Repo = 'repo',
    File = 'file',
    Dir = 'dir',
    Symbol = 'symbol',
}

export interface Suggestion {
    type: SuggestionType

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

export function createSuggestion(item: GQL.SearchSuggestion): Suggestion {
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: SuggestionType.Repo,
                title: item.uri,
                url: `/${item.uri}`,
                urlLabel: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts = [basename(item.repository.uri)]
            const dir = dirname(item.path)
            if (dir !== undefined && dir !== '.') {
                descriptionParts.push(`${dir}/`)
            }

            return {
                title: item.name,
                description: descriptionParts.join(' — '),
                type: item.isDirectory ? SuggestionType.Dir : SuggestionType.File,
                url: `/${item.repository.uri}/-/${item.isDirectory ? 'tree' : 'blob'}/${item.path}`,
                urlLabel: item.isDirectory ? 'go to dir' : 'go to file',
            }
        }
        case 'Symbol': {
            return {
                type: SuggestionType.Symbol,
                title: item.name,
                description: `${item.containerName || item.location.resource.path} – ${basename(
                    item.location.resource.repository.uri
                )}`,
                url: item.url,
                urlLabel: 'go to definition',
            }
        }
    }
}

const iconForType: { [key: string]: React.ComponentType<{ className: string }> } = {
    repo: RepoIcon,
    file: FileIcon,
    dir: FolderIcon,
    symbol: BookIcon,
}

export interface SuggestionProps {
    suggestion: Suggestion

    isSelected?: boolean

    /** Called when the user clicks on the suggestion */
    onClick?: () => void

    /** Get a reference to the HTML element for scroll management */
    liRef?: (ref: HTMLLIElement | null) => void
}

export const SuggestionItem = (props: SuggestionProps) => {
    const Icon = iconForType[props.suggestion.type]
    const suggestion = props.suggestion
    return (
        <li
            className={'suggestion2' + (props.isSelected ? ' suggestion2--selected' : '')}
            onMouseDown={props.onClick}
            ref={props.liRef}
        >
            <Icon className="icon-inline" />
            <div className="suggestion2__title">{suggestion.title}</div>
            <div className="suggestion2__description">{suggestion.description}</div>
            <div className="suggestion2__action" hidden={!props.isSelected}>
                <kbd>enter</kbd> {suggestion.urlLabel}
            </div>
        </li>
    )
}
