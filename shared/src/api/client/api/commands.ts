import { ProxyMarked, proxy, proxyMarker, Remote } from 'comlink'
import { Unsubscribable } from 'sourcegraph'
import { CommandRegistry } from '../services/command'
import { Subscription } from 'rxjs'
import { ProxySubscription } from './common'

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
        subscription.add(new ProxySubscription(run))
        return proxy(subscription)
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }
}
