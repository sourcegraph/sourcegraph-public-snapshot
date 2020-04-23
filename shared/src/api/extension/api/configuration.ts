import { Remote, ProxyMarked, proxyMarker } from '@sourcegraph/comlink'
import { ReplaySubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { SettingsCascade } from '../../../settings/settings'
import { ClientConfigurationAPI } from '../../client/api/configuration'

/**
 * @internal
 * @template C - The configuration schema.
 */
class ExtConfigurationSection<C extends object> implements sourcegraph.Configuration<C> {
    constructor(private proxy: Remote<ClientConfigurationAPI>, private data: C) {}

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
export interface ExtConfigurationAPI<C> extends ProxyMarked {
    $acceptConfigurationData(data: Readonly<SettingsCascade<C>>): void
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ExtConfiguration<C extends object> implements ExtConfigurationAPI<C>, ProxyMarked {
    public readonly [proxyMarker] = true

    /**
     * The settings data observable, assigned when the initial data is received from the client. Extensions should
     * never be able to call {@link ExtConfiguration}'s methods before the initial data is received.
     */
    private data?: Readonly<SettingsCascade<C>>

    // Buffer size of 1, so that sourcegraph.configuration:
    // - doesn't emit until initial settings have been received.
    // - emits immediately on subscription after initial settings have been received.
    public readonly changes = new ReplaySubject<void>(1)

    constructor(private proxy: Remote<ClientConfigurationAPI>) {}

    public $acceptConfigurationData(data: Readonly<SettingsCascade<C>>): void {
        this.data = Object.freeze(data)
        this.changes.next()
    }

    public get(): sourcegraph.Configuration<C> {
        if (!this.data) {
            throw new Error('unexpected internal error: settings data is not yet available')
        }
        return Object.freeze(new ExtConfigurationSection<C>(this.proxy, this.data.final))
    }
}
