import { ContextValues } from 'sourcegraph';
import { Connection } from '../../protocol/jsonrpc2/connection';
/** @internal */
export interface ClientContextAPI {
    $acceptContextUpdates(updates: ContextValues): void;
}
/** @internal */
export declare class ClientContext implements ClientContextAPI {
    private updateContext;
    private subscriptions;
    /**
     * Context keys set by this server. To ensure that context values are cleaned up, all context properties that
     * the server set are cleared upon deinitialization. This errs on the side of clearing too much (if another
     * server set the same context keys after this server, then those keys would also be cleared when this server's
     * client deinitializes).
     */
    private keys;
    constructor(connection: Connection, updateContext: (updates: ContextValues) => void);
    $acceptContextUpdates(updates: ContextValues): void;
    unsubscribe(): void;
}
