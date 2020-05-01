import { ProxyMarked, proxy, proxyMarker, Remote, releaseProxy } from 'comlink'
import { Unsubscribable } from 'sourcegraph'
import { CommandRegistry } from '../services/command'
import { Subscription } from 'rxjs'

/** @internal */
export interface ClientCommandsAPI extends ProxyMarked {
    $registerCommand(name: string, command: Remote<((...args: any) => any) & ProxyMarked>): Unsubscribable & ProxyMarked
    $executeCommand(command: string, args: any[]): Promise<any>
}

/** @internal */
export class ClientCommands implements ClientCommandsAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    constructor(private registry: CommandRegistry) {}

    public $registerCommand(command: string, run: Remote<(...args: any) => any>): Unsubscribable & ProxyMarked {
        const subscription = new Subscription()
        subscription.add(this.registry.registerCommand({ command, run }))
        subscription.add(() => run[releaseProxy]())
        return proxy(subscription)
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }
}
