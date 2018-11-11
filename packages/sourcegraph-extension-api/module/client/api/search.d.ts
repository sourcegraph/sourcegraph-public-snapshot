import { Connection } from 'src/protocol/jsonrpc2/connection';
import { TransformQuerySignature } from '../providers/queryTransformer';
import { FeatureProviderRegistry } from '../providers/registry';
/** @internal */
export interface SearchAPI {
    $registerQueryTransformer(id: number): void;
    $unregister(id: number): void;
}
/** @internal */
export declare class Search implements SearchAPI {
    private queryTransformerRegistry;
    private subscriptions;
    private registrations;
    private proxy;
    constructor(connection: Connection, queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>);
    $registerQueryTransformer(id: number): void;
    $unregister(id: number): void;
    unsubscribe(): void;
}
