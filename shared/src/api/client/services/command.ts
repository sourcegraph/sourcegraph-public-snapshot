import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import { WorkspaceEdit, SerializedWorkspaceEdit } from '../../types/workspaceEdit'
import { Diagnostic } from '@sourcegraph/extension-api-types'

/** A registered command in the command registry. */
export interface CommandEntry {
    /** The command ID (conventionally, e.g., "myextension.mycommand"). */
    command: string

    /** The function called to run the command and return an async value. */
    run: (...args: any[]) => Promise<any>
}

export interface ExecuteCommandParams {
    /**
     * The identifier of the actual command handler.
     */
    command: string

    /**
     * Arguments that the command should be invoked with.
     */
    arguments?: any[]
}

/** Manages and executes commands from all extensions. */
export class CommandRegistry {
    private entries = new BehaviorSubject<CommandEntry[]>([])

    public registerCommand(entry: CommandEntry): Unsubscribable {
        // Enforce uniqueness of command IDs.
        for (const e of this.entries.value) {
            if (e.command === entry.command) {
                throw new Error(`command is already registered: ${JSON.stringify(entry.command)}`)
            }
        }

        this.entries.next([...this.entries.value, entry])
        return {
            unsubscribe: () => {
                this.entries.next(this.entries.value.filter(e => e !== entry))
            },
        }
    }

    public executeCommand(params: ExecuteCommandParams): Promise<any> {
        return executeCommand(this.commandsSnapshot, params)
    }

    public async executeActionEditCommand(
        diagnostic: Diagnostic | null,
        params: ExecuteCommandParams
    ): Promise<SerializedWorkspaceEdit> {
        return WorkspaceEdit.fromJSON(
            await this.executeCommand({ ...params, arguments: [diagnostic, ...(params.arguments || [])] })
        )
    }

    /** All commands, emitted whenever the set of registered commands changed. */
    public readonly commands: Observable<CommandEntry[]> = this.entries

    /**
     * The current set of commands. Used by callers that do not need to react to commands being registered or
     * unregistered.
     */
    public get commandsSnapshot(): CommandEntry[] {
        return this.entries.value
    }
}

/**
 * Executes the command (in the commands list) specified in params.
 *
 * Most callers should use CommandRegistry's executeCommand method, which uses the registered commands.
 */
export function executeCommand(commands: CommandEntry[], params: ExecuteCommandParams): Promise<any> {
    const command = commands.find(c => c.command === params.command)
    if (!command) {
        throw new Error(`command not found: ${JSON.stringify(params.command)}`)
    }
    return Promise.resolve(command.run(...(params.arguments || [])))
}
