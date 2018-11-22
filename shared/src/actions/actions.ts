import H from 'history'
import { TextDocumentItem } from '../api/client/types/textDocument'
import { ContributableMenu, Contributions } from '../api/protocol'
import { ExtensionsContextProps } from '../context'
import { ControllerProps } from '../extensions/controller'

export interface ActionsProps extends ControllerProps, ExtensionsContextProps {
    menu: ContributableMenu
    scope?: TextDocumentItem
    actionItemClass?: string
    listClass?: string
    location: H.Location
}

export interface SearchFiltersProps extends ControllerProps, ExtensionsContextProps {
    scope?: TextDocumentItem
}

export interface ContributionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
