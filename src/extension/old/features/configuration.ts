import { BehaviorSubject } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import { MessageConnection } from '../../../jsonrpc2/connection'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationNotification,
    KeyPath,
} from '../../../protocol'
import { isEqual } from '../../../util'
import { Configuration, Observable } from '../api'

class ExtConfiguration<C> extends BehaviorSubject<C> implements Configuration<C>, Observable<C> {
    constructor(private rawConnection: MessageConnection, initial: ConfigurationCascade<C>) {
        super(initial.merged as C)

        rawConnection.onNotification(DidChangeConfigurationNotification.type, params => {
            this.next(params.configurationCascade.merged as C)
        })
    }

    public get<K extends keyof C>(key: K): C[K] | undefined {
        return this.value[key]
    }

    public watch<K extends keyof C>(...keys: K[]): Observable<Pick<C, K>> {
        return this.pipe(distinctUntilChanged((a, b) => keys.every(key => isEqual(a[key], b[key]))))
    }

    public update<K extends keyof C>(key: K, value: C[K] | undefined): Promise<void> {
        const path: KeyPath = [key as string | number]

        // Optimistically apply configuration update locally. If this diverges from the server's state, a
        // subsequent didChangeConfiguration notification will inform us.
        const cur = this.value
        this.next(setValueAtKeyPath(cur, path, value))

        return this.rawConnection.sendRequest(ConfigurationUpdateRequest.type, {
            path,
            value,
        } as ConfigurationUpdateParams)
    }

    public readonly [Symbol.observable] = () => this
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

/**
 * Creates the Sourcegraph extension API's {@link SourcegraphExtensionAPI#configuration} value.
 *
 * @param rawConnection The connection to the Sourcegraph API client.
 * @return The {@link SourcegraphExtensionAPI#configuration} value.
 */
export function createExtConfiguration<C>(
    rawConnection: MessageConnection,
    initial: ConfigurationCascade<C>
): Configuration<C> & Observable<C> {
    return new ExtConfiguration(rawConnection, initial)
}
