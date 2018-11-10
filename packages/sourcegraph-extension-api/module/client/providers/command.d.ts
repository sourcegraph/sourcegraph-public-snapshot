import { Observable, Unsubscribable } from 'rxjs';
/** A registered command in the command registry. */
export interface CommandEntry {
    /** The command ID (conventionally, e.g., "myextension.mycommand"). */
    command: string;
    /** The function called to run the command and return an async value. */
    run: (...args: any[]) => Promise<any>;
}
export interface ExecuteCommandParams {
    /**
     * The identifier of the actual command handler.
     */
    command: string;
    /**
     * Arguments that the command should be invoked with.
     */
    arguments?: any[];
}
/** Manages and executes commands from all extensions. */
export declare class CommandRegistry {
    private entries;
    registerCommand(entry: CommandEntry): Unsubscribable;
    executeCommand(params: ExecuteCommandParams): Promise<any>;
    /** All commands, emitted whenever the set of registered commands changed. */
    readonly commands: Observable<CommandEntry[]>;
    /**
     * The current set of commands. Used by callers that do not need to react to commands being registered or
     * unregistered.
     */
    readonly commandsSnapshot: CommandEntry[];
}
/**
 * Executes the command (in the commands list) specified in params.
 *
 * Most callers should use CommandRegistry's getHover method, which uses the registered commands.
 */
export declare function executeCommand(commands: CommandEntry[], params: ExecuteCommandParams): Promise<any>;
