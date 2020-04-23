import { ProxyMarked, proxy, proxyMarker } from '@sourcegraph/comlink'
import { Unsubscribable } from 'sourcegraph'
import { CommandRegistry } from '../services/command'

/** @internal */
export interface ClientCommandsAPI extends ProxyMarked {
    $registerCommand(name: string, command: (...args: any) => any): Unsubscribable & ProxyMarked
    $executeCommand(command: string, args: any[]): Promise<any>
}

/** @internal */
export class ClientCommands implements ClientCommandsAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    constructor(private registry: CommandRegistry) {}

    public $registerCommand(command: string, run: (...args: any) => any): Unsubscribable & ProxyMarked {
        return proxy(this.registry.registerCommand({ command, run }))
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }
}
