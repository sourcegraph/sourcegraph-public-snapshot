import { Unsubscribable } from 'rxjs';
import { ClientCommandsAPI } from '../../client/api/commands';
/** @internal */
export interface ExtCommandsAPI {
    $executeCommand(id: number, args: any[]): Promise<any>;
}
interface CommandEntry {
    command: string;
    callback: (...args: any[]) => any;
}
/** @internal */
export declare class ExtCommands implements ExtCommandsAPI {
    private proxy;
    private registrations;
    constructor(proxy: ClientCommandsAPI);
    /** Proxy method invoked by the client when the client wants to executes a command. */
    $executeCommand(id: number, args: any[]): Promise<any>;
    /**
     * Extension API method invoked directly when an extension wants to execute a command. It calls to the client
     * to execute the command because the desired command might be implemented on the client (or otherwise not in
     * this extension host).
     */
    executeCommand(command: string, args: any[]): Promise<any>;
    registerCommand(entry: CommandEntry): Unsubscribable;
}
export {};
