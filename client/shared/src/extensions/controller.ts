import type { Remote } from 'comlink'
import type { Unsubscribable } from 'rxjs'

import type { CommandEntry, ExecuteCommandParameters } from '../api/client/mainthread-api'
import type { FlatExtensionHostAPI } from '../api/contract'

export interface Controller extends Unsubscribable {
    /**
     * Executes the command (registered in the CommandRegistry) specified in params. If an error is thrown, the
     * error is returned *and* emitted on the {@link Controller#notifications} observable.
     *
     * All callers should execute commands using this method instead of calling
     * {@link sourcegraph:CommandRegistry#executeCommand} directly (to ensure errors are emitted as notifications).
     */
    executeCommand(parameters: ExecuteCommandParameters): Promise<any>

    registerCommand(entryToRegister: CommandEntry): Unsubscribable

    /**
     * Frees all resources associated with this client.
     */
    unsubscribe(): void

    extHostAPI: Promise<Remote<FlatExtensionHostAPI>>
}

/**
 * React props or state containing the client. There should be only a single client for the whole
 * application.
 */
export interface ExtensionsControllerProps<K extends keyof Controller = keyof Controller> {
    /**
     * The client, which is used to communicate with and manage extensions.
     */
    extensionsController: Pick<Controller, K> | null
}
export interface RequiredExtensionsControllerProps<K extends keyof Controller = keyof Controller> {
    extensionsController: Pick<Controller, K>
}
