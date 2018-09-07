import { Component } from 'sourcegraph/module/environment/environment'
import { ContributableMenu, Contributions } from 'sourcegraph/module/protocol'
import { ControllerProps } from '../../client/controller'
import { ExtensionsProps } from '../../context'
import { ConfigurationSubject, Settings } from '../../settings'

export interface ActionsProps<S extends ConfigurationSubject, C extends Settings>
    extends ControllerProps<S, C>,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
    scope?: Component
    actionItemClass?: string
    listClass?: string
}

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
