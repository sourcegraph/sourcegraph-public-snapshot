import * as sourcegraph from 'sourcegraph';
import { ClientConfigurationAPI } from '../../client/api/configuration';
import { ConfigurationCascade } from '../../protocol';
/**
 * @internal
 * @template C - The configuration schema.
 */
export interface ExtConfigurationAPI<C> {
    $acceptConfigurationData(data: Readonly<C>): Promise<void>;
}
/**
 * @internal
 * @template C - The configuration schema.
 */
export declare class ExtConfiguration<C extends ConfigurationCascade<any>> implements ExtConfigurationAPI<C> {
    private proxy;
    private data;
    constructor(proxy: ClientConfigurationAPI);
    $acceptConfigurationData(data: Readonly<C>): Promise<void>;
    get(): sourcegraph.Configuration<C>;
    subscribe(next: () => void): sourcegraph.Unsubscribable;
}
