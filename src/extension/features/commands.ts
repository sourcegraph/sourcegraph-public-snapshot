import { Subscription } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { CommandRegistry } from '../../environment/providers/command'
import {
    ExecuteCommandRegistrationOptions,
    ExecuteCommandRequest,
    RegistrationParams,
    RegistrationRequest,
    UnregistrationParams,
    UnregistrationRequest,
} from '../../protocol'
import { Commands, CXP } from '../api'

/**
 * Creates the CXP extension API's {@link CXP#commands} value.
 *
 * @param ext The CXP extension API handle.
 * @return The {@link CXP#commands} value.
 */
export function createExtCommands(ext: Pick<CXP<any>, 'rawConnection'>): Commands {
    // TODO!(sqs): move CommandRegistry to somewhere general since it's now used by the controller AND extension
    const commandRegistry = new CommandRegistry()
    ext.rawConnection.onRequest(ExecuteCommandRequest.type, params => commandRegistry.executeCommand(params))
    return {
        register: (command: string, run: (...args: any[]) => Promise<any>): Subscription => {
            const subscription = new Subscription()

            const id = uuidv4()
            subscription.add(commandRegistry.registerCommand({ command, run }))
            ext.rawConnection
                .sendRequest(RegistrationRequest.type, {
                    registrations: [
                        {
                            id,
                            method: ExecuteCommandRequest.type.method,
                            registerOptions: { commands: [command] } as ExecuteCommandRegistrationOptions,
                        },
                    ],
                } as RegistrationParams)
                .catch(err => console.error(err))

            subscription.add(() => {
                ext.rawConnection
                    .sendRequest(UnregistrationRequest.type, {
                        unregisterations: [{ id, method: ExecuteCommandRequest.type.method }],
                    } as UnregistrationParams)
                    .catch(err => console.error(err))
            })
            return subscription
        },
    }
}
