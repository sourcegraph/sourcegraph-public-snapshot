import H from 'history'
import { TextDocumentItem } from '../api/client/types/textDocument'
import { ContributableMenu, Contributions } from '../api/protocol'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'

export interface ActionsProps extends ExtensionsControllerProps, PlatformContextProps {
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
