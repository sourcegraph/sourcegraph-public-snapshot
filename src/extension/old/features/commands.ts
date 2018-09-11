import { Subscription } from 'rxjs'
import { CommandRegistry } from '../../../client/providers/command'
import {
    ExecuteCommandRegistrationOptions,
    ExecuteCommandRequest,
    RegistrationParams,
    RegistrationRequest,
    UnregistrationParams,
    UnregistrationRequest,
} from '../../../protocol'
import { Connection } from '../../../protocol/jsonrpc2/connection'
import { idSequence } from '../../../util'
import { Commands } from '../api'

/**
 * Creates the Sourcegraph extension API's {@link SourcegraphExtensionAPI#commands} value.
 *
 * @param rawConnection The connection to the Sourcegraph API client.
 * @return The {@link Creates the Sourcegraph extension API extension API's#commands} value.
 */
export function createExtCommands(rawConnection: Connection): Commands {
    // TODO: move CommandRegistry to somewhere general since it's now used by the controller AND extension
    const commandRegistry = new CommandRegistry()
    rawConnection.onRequest(ExecuteCommandRequest.type, params => commandRegistry.executeCommand(params))
    return {
        register: (command: string, run: (...args: any[]) => Promise<any>): Subscription => {
            const subscription = new Subscription()

            const id = idSequence()
            subscription.add(commandRegistry.registerCommand({ command, run }))
            rawConnection
                .sendRequest(RegistrationRequest.type, {
                    registrations: [
                        {
                            id,
                            method: ExecuteCommandRequest.type,
                            registerOptions: { commands: [command] } as ExecuteCommandRegistrationOptions,
                        },
                    ],
                } as RegistrationParams)
                .catch(err => console.error(err))

            subscription.add(() => {
                rawConnection
                    .sendRequest(UnregistrationRequest.type, {
                        unregisterations: [{ id, method: ExecuteCommandRequest.type }],
                    } as UnregistrationParams)
                    .catch(err => console.error(err))
            })
            return subscription
        },
    }
}
