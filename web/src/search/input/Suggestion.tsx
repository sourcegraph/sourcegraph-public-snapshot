import * as React from 'react'
import { RepositoryIcon } from '../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import * as GQL from '../../../../shared/src/graphql/schema'
import { SymbolIcon } from '../../../../shared/src/symbols/SymbolIcon'

interface BaseSuggestion {
    title: string
    description?: string
}

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

interface Filters extends BaseSuggestion {
    type: SuggestionTypes.filters
}
interface Repo extends BaseSuggestion {
    type: SuggestionTypes.repo
}
interface Repogroup extends BaseSuggestion {
    type: SuggestionTypes.repogroup
}
interface Repohasfile extends BaseSuggestion {
    type: SuggestionTypes.repohasfile
}
interface Repohascommitafter extends BaseSuggestion {
    type: SuggestionTypes.repohascommitafter
}
interface File extends BaseSuggestion {
    type: SuggestionTypes.file
}
interface Type extends BaseSuggestion {
    type: SuggestionTypes.type
}
interface Case extends BaseSuggestion {
    type: SuggestionTypes.case
}
interface Lang extends BaseSuggestion {
    type: SuggestionTypes.lang
}
interface Fork extends BaseSuggestion {
    type: SuggestionTypes.fork
}
interface Archived extends BaseSuggestion {
    type: SuggestionTypes.archived
}
interface Count extends BaseSuggestion {
    type: SuggestionTypes.count
}
interface Timeout extends BaseSuggestion {
    type: SuggestionTypes.timeout
}

export type Suggestion =
    | Filters
    | Repo
    | Repogroup
    | Repohasfile
    | Repohascommitafter
    | File
    | Type
    | Case
    | Lang
    | Fork
    | Archived
    | Count
    | Timeout

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
