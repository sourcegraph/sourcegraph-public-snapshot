import { ProxyValue, proxy, proxyMarker } from '@sourcegraph/comlink'
import { Unsubscribable } from 'sourcegraph'
import { CommandRegistry } from '../services/command'

/** @internal */
export interface ClientCommandsAPI extends ProxyValue {
    $registerCommand(name: string, command: (...args: any) => any): Unsubscribable & ProxyValue
    $executeCommand(command: string, args: any[]): Promise<any>
}

/** @internal */
export class ClientCommands implements ClientCommandsAPI, ProxyValue {
    public readonly [proxyMarker] = true

    constructor(private registry: CommandRegistry) {}

    public $registerCommand(command: string, run: (...args: any) => any): Unsubscribable & ProxyValue {
        return proxy(this.registry.registerCommand({ command, run }))
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }
}
