import { BehaviorSubject, Observable, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { Emitter, Event } from '../../jsonrpc2/events'
import { InitializeParams, ServerCapabilities } from '../../protocol'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    KeyPath,
    Settings,
} from '../../protocol/configuration'
import { isEqual } from '../../util'
import { Connection } from '../server'
import { Remote } from './common'

/**
 * A proxy for the client's configuration.
 *
 * @template C settings type
 */
export class RemoteConfiguration<C extends Settings> implements Remote, Unsubscribable {
    private subscription = new Subscription()
    private _connection?: Connection
    private _configurationCascade = new BehaviorSubject<ConfigurationCascade<C>>({ merged: {} as C })
    private onChange = new Emitter<void>()

    /**
     * An observable of the configuration that emits whenever it changes (when the extension itself updates the
     * configuration, or when the extension receives a workspace/didChangeConfiguration notification from the
     * client).
     *
     * The current merged configuration is available in the `value` property of the returned object, for callers
     * that want to access it directly without subscribing to the observable.
     */
    public get configuration(): Observable<Readonly<C>> & { value: Readonly<C> } {
        const o: Observable<C> & { value: C } = Object.create(
            this._configurationCascade.pipe(
                map(({ merged }) => merged),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
        o.value = this._configurationCascade.value.merged
        return o
    }

    /**
     * Emits when the configuration changes (when the extension itself updates the configuration, or when the
     * extension receives a workspace/didChangeConfiguration notification from the client).
     */
    public readonly onDidChangeConfiguration: Event<void> = this.onChange.event

    public attach(connection: Connection): void {
        this._connection = connection

        // Listen for `workspace/didChangeConfiguration` notifications from the client.
        this.subscription.add(
            connection.onDidChangeConfiguration(params => {
                this._configurationCascade.next(params.configurationCascade as ConfigurationCascade<C>)
                this.onChange.fire(void 0)
            })
        )
    }

    public get connection(): Connection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public initialize(params: InitializeParams): void {
        this._configurationCascade.next(params.configurationCascade as ConfigurationCascade<C>)
        this.onChange.fire(void 0)
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    /**
     * Updates the configuration setting at the given key path to the given value. The local merged configuration
     * is immediately updated to reflect the change (optimistically, even before the server acknowledges the
     * update).
     *
     * Implementation: sends a configuration/update notification to the client.
     *
     * @param path the key path of the configuration setting to update
     * @param value the value to insert
     */
    public updateConfiguration(path: KeyPath, value: any): Promise<void> {
        // Optimistically apply configuration update locally. If this diverges from the server's state, a
        // subsequent didChangeConfiguration notification will inform us.
        const cur = this._configurationCascade.value
        this._configurationCascade.next({ ...cur, merged: setValueAtKeyPath(cur.merged, path, value) })
        this.onChange.fire(void 0)

        return this.connection.sendRequest(ConfigurationUpdateRequest.type, {
            path,
            value,
        } as ConfigurationUpdateParams)
    }

    public unsubscribe(): void {
        this.subscription.unsubscribe()
    }
}

/**
 * Returns the source object with the given value inserted in the location specified by the key path. The source
 * object is not modified. The key path indexes into the object successively for each element in the key path.
 *
 * If the value is `undefined`, the value at the key path is removed.
 *
 * This must behave identically to {@link module:jsonc-parser.setProperty}.
 */
export function setValueAtKeyPath(source: any, path: KeyPath, value: any): any {
    if (path.length === 0) {
        // Overwrite entire value.
        return value
    }

    const root = [source]
    let prev: any = root // maintain an lvalue that we can assign to
    for (const [i, key] of path.entries()) {
        const last = i === path.length - 1
        const prevKey = i === 0 ? 0 : path[i - 1]
        if (typeof key === 'string') {
            if (last) {
                if (value === undefined) {
                    prev[prevKey] = { ...prev[prevKey] }
                    delete prev[prevKey][key]
                } else if (
                    prev[prevKey] !== null &&
                    typeof prev[prevKey] === 'object' &&
                    !Array.isArray(prev[prevKey])
                ) {
                    prev[prevKey] = { ...prev[prevKey], [key]: value }
                } else {
                    prev[prevKey] = { [key]: value }
                }
            } else {
                prev[prevKey] =
                    prev[prevKey] !== null && typeof prev[prevKey] === 'object' && !Array.isArray(prev[prevKey])
                        ? { ...prev[prevKey] }
                        : {}
            }
        } else if (typeof key === 'number') {
            if (last) {
                const index = key === -1 ? prev[prevKey].length : key
                const head = Array.isArray(prev[prevKey]) ? prev[prevKey].slice(0, index) : []
                const tail = Array.isArray(prev[prevKey]) ? prev[prevKey].slice(index + 1) : []
                if (value === undefined) {
                    prev[prevKey] = [...head, ...tail]
                } else {
                    prev[prevKey] = [...head, value, ...tail]
                }
            } else {
                prev[prevKey] = Array.isArray(prev[prevKey]) ? [...prev[prevKey]] : []
            }
        } else {
            throw new Error(`invalid key in key path: ${key} (full key path: ${JSON.stringify(path)}`)
        }
        prev = prev[prevKey]
    }
    return root[0]
}
