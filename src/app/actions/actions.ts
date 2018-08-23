import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { ExtensionsProps } from '../../context'
import { CXPControllerProps } from '../../cxp/controller'
import { ConfigurationSubject, Settings } from '../../settings'

export interface ActionsProps<S extends ConfigurationSubject, C extends Settings>
    extends CXPControllerProps<S, C>,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
    actionItemClass?: string
    listClass?: string
}

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
