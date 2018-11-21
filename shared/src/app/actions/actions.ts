import H from 'history'
import { ExtensionsProps } from '../../../../web/src/extensions/ExtensionsClientCommonContext'
import { TextDocumentItem } from '../../api/client/types/textDocument'
import { ContributableMenu, Contributions } from '../../api/protocol'
import { ControllerProps } from '../../client/controller'
import { ExtensionsContextProps } from '../../context'
import { Settings, SettingsSubject } from '../../settings'

export interface ActionsProps extends ControllerProps, ExtensionsContextProps {
    menu: ContributableMenu
    scope?: TextDocumentItem
    actionItemClass?: string
    listClass?: string
    location: H.Location
}

export interface SearchFiltersProps<S extends SettingsSubject, C extends Settings>
    extends ControllerProps<S, C>,
        ExtensionsProps<S, C> {
    scope?: TextDocumentItem
}

export interface ContributionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
