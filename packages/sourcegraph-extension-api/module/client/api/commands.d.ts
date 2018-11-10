import { Connection } from '../../protocol/jsonrpc2/connection';
import { CommandRegistry } from '../providers/command';
/** @internal */
export interface ClientCommandsAPI {
    $unregister(id: number): void;
    $registerCommand(id: number, command: string): void;
    $executeCommand(command: string, args: any[]): Promise<any>;
}
/** @internal */
export declare class ClientCommands implements ClientCommandsAPI {
    private registry;
    private subscriptions;
    private registrations;
    private proxy;
    constructor(connection: Connection, registry: CommandRegistry);
    $unregister(id: number): void;
    $registerCommand(id: number, command: string): void;
    $executeCommand(command: string, args: any[]): Promise<any>;
    unsubscribe(): void;
}
