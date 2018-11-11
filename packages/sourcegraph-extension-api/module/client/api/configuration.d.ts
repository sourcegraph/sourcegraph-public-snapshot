import { Observable } from 'rxjs';
import { ConfigurationUpdateParams } from '../../protocol';
import { Connection } from '../../protocol/jsonrpc2/connection';
/** @internal */
export interface ClientConfigurationAPI {
    $acceptConfigurationUpdate(params: ConfigurationUpdateParams): Promise<void>;
}
/**
 * @internal
 * @template C - The configuration schema.
 */
export declare class ClientConfiguration<C> implements ClientConfigurationAPI {
    private updateConfiguration;
    private subscriptions;
    private proxy;
    constructor(connection: Connection, environmentConfiguration: Observable<C>, updateConfiguration: (params: ConfigurationUpdateParams) => Promise<void>);
    $acceptConfigurationUpdate(params: ConfigurationUpdateParams): Promise<void>;
    unsubscribe(): void;
}
