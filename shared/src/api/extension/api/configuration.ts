import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { ReplaySubject } from 'rxjs'
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

const BUFFER_SIZE = 1

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ExtConfiguration<C extends object> extends ReplaySubject<void>
    implements ExtConfigurationAPI<C>, ProxyValue {
    public readonly [proxyValueSymbol] = true

    /**
     * The settings data observable, assigned when the initial data is received from the client. Extensions should
     * never be able to call {@link ExtConfiguration}'s methods before the initial data is received.
     */
    private data?: Readonly<SettingsCascade<C>>

    constructor(private proxy: ProxyResult<ClientConfigurationAPI>) {
        // Call super() with a buffer size of 1, so that sourcegraph.configuration:
        // - doesn't emit until initial settings have been received.
        // - emits immediately on subscription after initial settings have been received.
        super(BUFFER_SIZE)
    }

    public $acceptConfigurationData(data: Readonly<SettingsCascade<C>>): void {
        this.data = Object.freeze(data)
        this.next()
    }

    public get(): sourcegraph.Configuration<C> {
        if (!this.data) {
            throw new Error('unexpected internal error: settings data is not yet available')
        }
        return Object.freeze(new ExtConfigurationSection<C>(this.proxy, this.data.final))
    }
}
