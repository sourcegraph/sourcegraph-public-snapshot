import H from 'history'
import { TextDocumentItem } from 'sourcegraph/module/client/types/textDocument'
import { ContributableMenu, Contributions } from 'sourcegraph/module/protocol'
import { ControllerProps } from '../../client/controller'
import { ExtensionsProps } from '../../context'
import { ConfigurationSubject, Settings } from '../../settings'

export interface ActionsProps<S extends ConfigurationSubject, C extends Settings>
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
