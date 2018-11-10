import { Subscription } from 'rxjs'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtCommandsAPI } from '../../extension/api/commands'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { CommandRegistry } from '../providers/command'
import { SubscriptionMap } from './common'

/** @internal */
export interface ClientCommandsAPI {
    $unregister(id: number): void
    $registerCommand(id: number, command: string): void
    $executeCommand(command: string, args: any[]): Promise<any>
}

/** @internal */
export class ClientCommands implements ClientCommandsAPI {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtCommandsAPI

    constructor(connection: Connection, private registry: CommandRegistry) {
        this.subscriptions.add(this.registrations)

        this.proxy = createProxyAndHandleRequests('commands', connection, this)
    }

    public $unregister(id: number): void {
        this.registrations.remove(id)
    }

    public $registerCommand(id: number, command: string): void {
        this.registrations.add(
            id,
            this.registry.registerCommand({
                command,
                run: (...args: any[]): any => this.proxy.$executeCommand(id, args),
            })
        )
    }

    public $executeCommand(command: string, args: any[]): Promise<any> {
        return this.registry.executeCommand({ command, arguments: args })
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
