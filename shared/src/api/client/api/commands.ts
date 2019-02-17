import { ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { Unsubscribable } from 'sourcegraph'
import { CommandRegistry } from '../services/command'

/** @internal */
export interface ClientCommandsAPI extends ProxyValue {
    $registerCommand(name: string, command: (...args: any) => any): Unsubscribable & ProxyValue
    $executeCommand(command: string, args: any[]): Promise<any>
}

/** @internal */
export class ClientCommands implements ClientCommandsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(private registry: CommandRegistry) {}

    public $registerCommand(command: string, run: (...args: any) => any): Unsubscribable & ProxyValue {
        return proxyValue(this.registry.registerCommand({ command, run }))
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }
}
