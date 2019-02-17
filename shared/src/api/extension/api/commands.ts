import { ProxyResult, proxyValue } from 'comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import { ClientCommandsAPI } from '../../client/api/commands'

interface CommandEntry {
    command: string
    callback: (...args: any[]) => any
}

/** @internal */
export class ExtCommands {
    constructor(private proxy: ProxyResult<ClientCommandsAPI>) {}

    /**
     * Extension API method invoked directly when an extension wants to execute a command. It calls to the client
     * to execute the command because the desired command might be implemented on the client (or otherwise not in
     * this extension host).
     */
    public executeCommand(command: string, args: any[]): Promise<any> {
        return this.proxy.$executeCommand(command, args)
    }

    public registerCommand(entry: CommandEntry): Unsubscribable {
        const subscription = new Subscription()
        // tslint:disable-next-line: no-floating-promises
        this.proxy.$registerCommand(entry.command, proxyValue(entry.callback)).then(s => subscription.add(s))
        return subscription
    }
}
