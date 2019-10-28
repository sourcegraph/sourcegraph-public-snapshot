import * as React from 'react'
import { RepositoryIcon } from '../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import * as GQL from '../../../../shared/src/graphql/schema'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'
import { LanguageIcon } from '../../../../shared/src/components/languageIcons'

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
}

export interface Suggestion {
    title: string
    description?: string
    type: SuggestionTypes
}

interface SuggestionIconProps {
    suggestion: Suggestion
    className?: string
}

const SuggestionIcon: React.FunctionComponent<SuggestionIconProps> = ({ suggestion, ...passThru }) => {
    switch (suggestion.type) {
        case SuggestionTypes.repo:
            return <RepositoryIcon {...passThru} />
        case SuggestionTypes.file:
            return <SymbolIcon kind={GQL.SymbolKind.FILE} {...passThru} />
        case SuggestionTypes.lang:
            return <LanguageIcon language={suggestion.title} {...passThru} />
        default:
            return null // TODO: handle lang suggestions in RFC 14 frontend PR.
    }
}

interface SuggestionProps {
    suggestion: Suggestion

    isSelected?: boolean

    /** Called when the user clicks on the suggestion */
    onClick?: () => void

    /** Get a reference to the HTML element for scroll management */
    ref?: (ref: HTMLLIElement | null) => void
}

export const SuggestionItem: React.FunctionComponent<SuggestionProps> = ({ suggestion, isSelected, ...props }) => (
    <li className={'suggestion' + (isSelected ? ' suggestion--selected' : '')} {...props}>
        <SuggestionIcon className="icon-inline suggestion__icon" suggestion={suggestion} />
        <div className="suggestion__title">{suggestion.title}</div>
        <div className="suggestion__description">{suggestion.description}</div>
    </li>
)
