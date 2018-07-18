import { Subscription } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { CommandRegistry } from '../../environment/providers/command'
import { MessageType as RPCMessageType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    ExecuteCommandParams,
    ExecuteCommandRegistrationOptions,
    ExecuteCommandRequest,
    ServerCapabilities,
} from '../../protocol'
import { Client } from '../client'
import { DynamicFeature, ensure, RegistrationData } from './common'

/**
 * Support for the commands executed on the server by the client (workspace/executeCommand requests to the server).
 */
export class ExecuteCommandFeature implements DynamicFeature<ExecuteCommandRegistrationOptions> {
    private commands = new Map<string, Subscription>()

    constructor(private client: Client, private registry: CommandRegistry) {}

    public get messages(): RPCMessageType {
        return ExecuteCommandRequest.type
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        ensure(ensure(capabilities, 'workspace')!, 'executeCommand')!.dynamicRegistration = true
    }

    public initialize(capabilities: ServerCapabilities): void {
        if (!capabilities.executeCommandProvider) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { ...capabilities.executeCommandProvider },
        })
    }

    public register(_message: RPCMessageType, data: RegistrationData<ExecuteCommandRegistrationOptions>): void {
        const sub = new Subscription()
        for (const command of data.registerOptions.commands) {
            sub.add(
                this.registry.registerCommand({
                    command,
                    run: (...args: any[]): Promise<any> =>
                        this.client.sendRequest(ExecuteCommandRequest.type, {
                            command,
                            arguments: args,
                        } as ExecuteCommandParams),
                })
            )
        }
        this.commands.set(data.id, sub)
    }

    public unregister(id: string): void {
        const sub = this.commands.get(id)
        if (sub) {
            sub.unsubscribe()
        }
        this.commands.delete(id)
    }

    public unregisterAll(): void {
        for (const sub of this.commands.values()) {
            sub.unsubscribe()
        }
        this.commands.clear()
    }
}
