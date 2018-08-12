import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { ExtensionsProps } from '../../context'
import { CXPControllerProps } from '../../cxp/controller'
import { ConfigurationCascade, ConfigurationSubject } from '../../settings'

export interface ActionsProps<S extends ConfigurationSubject, C extends ConfigurationCascade<S>>
    extends CXPControllerProps<S, C>,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
}

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
