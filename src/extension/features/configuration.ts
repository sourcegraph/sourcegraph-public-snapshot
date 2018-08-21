import { BehaviorSubject } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationNotification,
} from '../../protocol'
import { isEqual } from '../../util'
import { Configuration, CXP, Observable } from '../api'

class ExtConfiguration<C> extends BehaviorSubject<C> implements Configuration<C>, Observable<C> {
    constructor(private ext: Pick<CXP<C>, 'rawConnection'>, initial: ConfigurationCascade<C>) {
        super(initial.merged as C)

        ext.rawConnection.onNotification(DidChangeConfigurationNotification.type, params => {
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
        // TODO!(sqs): update config locally here
        return this.ext.rawConnection.sendRequest(ConfigurationUpdateRequest.type, {
            path: [key],
            value,
        } as ConfigurationUpdateParams)
    }

    public readonly [Symbol.observable] = () => this
}

/**
 * Creates the CXP extension API's {@link CXP#configuration} value.
 *
 * @param ext The CXP extension API handle.
 * @return The {@link CXP#configuration} value.
 */
export function createExtConfiguration<C>(
    ext: Pick<CXP<C>, 'rawConnection'>,
    initial: ConfigurationCascade<C>
): Configuration<C> & Observable<C> {
    return new ExtConfiguration(ext, initial)
}
