import { ProxyResult, ProxyValue, proxyMarker } from '@sourcegraph/comlink'
import { BehaviorSubject, PartialObserver, Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { SettingsCascade } from '../../../settings/settings'
import { ClientConfigurationAPI } from '../../client/api/configuration'

/**
 * @internal
 * @template C - The configuration schema.
 */
class ExtConfigurationSection<C extends object> implements sourcegraph.Configuration<C> {
    constructor(private proxy: ProxyResult<ClientConfigurationAPI>, private data: C) {}

    public get<K extends keyof C>(key: K): C[K] | undefined {
        return this.data[key]
    }

    public async update<K extends keyof C>(key: K, value: C[K] | undefined): Promise<void> {
        // Cast `key` to `string | number` (i.e., eliminate `symbol`). We could use `Extract<keyof
        // C, string | number` in the sourcegraph.d.ts type signature, but that would add useless
        // complexity.
        await this.proxy.$acceptConfigurationUpdate({ path: [key as string | number], value })
    }

    public get value(): Readonly<C> {
        return this.data
    }

    public toJSON(): any {
        return this.data
    }
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export interface ExtConfigurationAPI<C> extends ProxyValue {
    $acceptConfigurationData(data: Readonly<SettingsCascade<C>>): void
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ExtConfiguration<C extends object> implements ExtConfigurationAPI<C>, ProxyValue {
    public readonly [proxyMarker] = true

    /**
     * The settings data observable, assigned when the initial data is received from the client. Extensions should
     * never be able to call {@link ExtConfiguration}'s methods before the initial data is received.
     */
    private data?: BehaviorSubject<Readonly<SettingsCascade<C>>>

    constructor(private proxy: ProxyResult<ClientConfigurationAPI>) {}

    public $acceptConfigurationData(data: Readonly<SettingsCascade<C>>): void {
        if (!this.data) {
            this.data = new BehaviorSubject(data)
        } else {
            this.data.next(Object.freeze(data))
        }
    }

    private getData(): BehaviorSubject<Readonly<SettingsCascade<C>>> {
        if (!this.data) {
            throw new Error('unexpected internal error: settings data is not yet available')
        }
        return this.data
    }

    public get(): sourcegraph.Configuration<C> {
        return Object.freeze(new ExtConfigurationSection<C>(this.proxy, this.getData().value.final))
    }

    public subscribe(observer?: PartialObserver<C>): Unsubscribable
    public subscribe(next?: (value: C) => void, error?: (error: any) => void, complete?: () => void): Unsubscribable
    public subscribe(...args: any[]): sourcegraph.Unsubscribable {
        return this.getData().subscribe(...args)
    }
}
