import FileIcon from '@sourcegraph/icons/lib/File'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as React from 'react'
import { buildSearchURLQuery, SearchOptions } from './index'

export const enum SuggestionType {
    Repo = 'repo',
    File = 'file',
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

export function createSuggestion(item: GQL.SearchSuggestion2, options: SearchOptions): Suggestion {
    const searchQuerystring = buildSearchURLQuery(options)
    switch (item.__typename) {
        case 'Repository': {
            return {
                type: SuggestionType.Repo,
                title: item.uri,
                url: `/${item.uri}?${searchQuerystring}`,
                urlLabel: 'go to repository',
            }
        }
        case 'File': {
            const descriptionParts = [basename(item.repository.uri)]
            const dir = dirname(item.name)
            if (dir !== undefined) {
                descriptionParts.push(`${dir}/`)
            }

            return {
                type: SuggestionType.File,
                title: basename(item.name),
                description: descriptionParts.join(' — '),
                url: `/${item.repository.uri}/-/blob/${item.name}?${searchQuerystring}`,
                urlLabel: 'go to file',
            }
        }
    }
}

const iconForType: { [key: string]: React.ComponentType<{ className: string }> } = {
    repo: RepoIcon,
    file: FileIcon,
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

function dirname(path: string): string | undefined {
    return path.split('/').slice(-2, -1)[0]
}

function basename(path: string): string {
    return path.split('/').slice(-1)[0]
}
