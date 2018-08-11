import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { ExtensionsProps } from '../../context'
import { Settings } from '../../copypasta'
import { CXPControllerProps } from '../../cxp/controller'
import { ConfigurationSubject } from '../../settings'

export interface ActionsProps<S extends ConfigurationSubject, C = Settings>
    extends CXPControllerProps,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
}

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
