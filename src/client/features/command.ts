import { BehaviorSubject, from, Observable, Subscription, TeardownLogic, throwError } from 'rxjs'
import { first, map } from 'rxjs/operators'
import * as uuidv4 from 'uuid/v4'
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

export type ExecuteCommandSignature = (params: ExecuteCommandParams) => Promise<any>

interface CommandEntry {
    /** The command ID (conventionally, e.g., "myextension.mycommand"). */
    command: string

    run: (...args: any[]) => Promise<any>
}

export class CommandRegistry {
    private entries = new BehaviorSubject<CommandEntry[]>([])

    public registerCommand(entry: CommandEntry): TeardownLogic {
        // Enforce uniqueness of command IDs.
        for (const e of this.entries.value) {
            if (e.command === entry.command) {
                throw new Error(`command is already registered: ${JSON.stringify(entry.command)}`)
            }
        }

        this.entries.next([...this.entries.value, entry])
        return () => {
            this.entries.next(this.entries.value.filter(e => e !== entry))
        }
    }

    public executeCommand(params: ExecuteCommandParams): Promise<any> {
        return this.commandsSnapshot
            .pipe(
                map(commands => {
                    const command = commands.find(c => c.command === params.command)
                    if (!command) {
                        return throwError(new Error(`command not found: ${JSON.stringify(params.command)}`))
                    }
                    return from(command.run(...(params.arguments || [])))
                })
            )
            .toPromise()
    }

    /** All commands, emitted whenever the set of registered commands changed. */
    public readonly commands: Observable<CommandEntry[]> = this.entries

    /**
     * The current set of commands. Used by callers that do not need to react to commands being registered or
     * unregistered.
     */
    public readonly commandsSnapshot = this.commands.pipe(first())
}

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

    public unsubscribe(): void {
        for (const sub of this.commands.values()) {
            sub.unsubscribe()
        }
        this.commands.clear()
    }
}
