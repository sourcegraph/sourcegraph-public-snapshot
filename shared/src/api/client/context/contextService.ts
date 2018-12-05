import { BehaviorSubject, NextObserver, Subscribable } from 'rxjs'
import { PlatformContext } from '../../../platform/context'
import { Context } from './context'

/**
 * The context service owns the context data for the context, which consists of arbitrary key-value pairs that
 * describe application state.
 */
export interface ContextService {
    /**
     * The context data.
     */
    readonly data: Subscribable<Context> & { value: Context } & NextObserver<Context>
}

/** Create a {@link ContextService} instance. */
export function createContextService({
    clientApplication,
}: Pick<PlatformContext, 'clientApplication'>): ContextService {
    return {
        data: new BehaviorSubject<Context>({ 'clientApplication.isSourcegraph': clientApplication === 'sourcegraph' }),
    }
}
