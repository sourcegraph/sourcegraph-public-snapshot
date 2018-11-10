import H from 'history'
import { TextDocumentItem } from '../../api/client/types/textDocument'
import { ContributableMenu, Contributions } from '../../api/protocol'
import { ControllerProps } from '../../client/controller'
import { ExtensionsProps } from '../../context'
import { Settings, SettingsSubject } from '../../settings'

export interface ActionsProps<S extends SettingsSubject, C extends Settings>
    extends ControllerProps<S, C>,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
    scope?: TextDocumentItem
    actionItemClass?: string
    listClass?: string
    location: H.Location
}

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
