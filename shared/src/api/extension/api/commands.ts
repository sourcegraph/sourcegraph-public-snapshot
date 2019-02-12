import { ProxyValue, proxyValueSymbol } from 'comlink'
import { Unsubscribable } from 'rxjs'
import { ClientCommandsAPI } from '../../client/api/commands'
import { ProviderMap } from './common'

/** @internal */
export interface ExtCommandsAPI {
    $executeCommand(id: number, args: any[]): Promise<any>
}

interface CommandEntry {
    command: string
    callback: (...args: any[]) => any
}

/** @internal */
export class ExtCommands implements ExtCommandsAPI, Unsubscribable, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private registrations = new ProviderMap<CommandEntry>(id => this.proxy.$unregister(id))

    constructor(private proxy: ClientCommandsAPI) {}

    /** Proxy method invoked by the client when the client wants to executes a command. */
    public $executeCommand(id: number, args: any[]): Promise<any> {
        const { callback } = this.registrations.get<CommandEntry>(id)
        return Promise.resolve(callback(...args))
    }

    /**
     * Extension API method invoked directly when an extension wants to execute a command. It calls to the client
     * to execute the command because the desired command might be implemented on the client (or otherwise not in
     * this extension host).
     */
    public executeCommand(command: string, args: any[]): Promise<any> {
        return this.proxy.$executeCommand(command, args)
    }

    public registerCommand(entry: CommandEntry): Unsubscribable {
        const { id, subscription } = this.registrations.add(entry)
        this.proxy.$registerCommand(id, entry.command)
        return subscription
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
