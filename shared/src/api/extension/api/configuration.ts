import { BehaviorSubject } from 'rxjs'
import { filter } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ClientConfigurationAPI } from '../../client/api/configuration'
import { SettingsCascade } from '../../protocol'

/**
 * @internal
 * @template C - The configuration schema.
 */
class ExtConfigurationSection<C extends object> implements sourcegraph.Configuration<C> {
    constructor(private proxy: ClientConfigurationAPI, private data: C) {}

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
export interface ExtConfigurationAPI<C> {
    $acceptConfigurationData(data: Readonly<C>): Promise<void>
}

/**
 * @internal
 * @template C - The configuration schema.
 */
export class ExtConfiguration<C extends SettingsCascade<any>> implements ExtConfigurationAPI<C> {
    private data: BehaviorSubject<Readonly<C>>

    constructor(private proxy: ClientConfigurationAPI, initialData: Readonly<C>) {
        this.data = new BehaviorSubject<Readonly<C>>(initialData)
    }

    public $acceptConfigurationData(data: Readonly<C>): Promise<void> {
        this.data.next(Object.freeze(data))
        return Promise.resolve()
    }

    public get(): sourcegraph.Configuration<C> {
        return Object.freeze(new ExtConfigurationSection<C>(this.proxy, this.data.value.final))
    }

    public subscribe(next: () => void): sourcegraph.Unsubscribable {
        // Do not emit until the configuration is available.
        return this.data.pipe(filter(data => data !== null)).subscribe(next)
    }
}
