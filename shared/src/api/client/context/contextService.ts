import { BehaviorSubject } from 'rxjs'
import { PlatformContext } from '../../../platform/context'
import { BehaviorSubjectLike } from '../services/util'
import { Context } from './context'

/**
 * The context service owns the context data for the context, which consists of arbitrary key-value pairs that
 * describe application state.
 */
export interface ContextService {
    /**
     * The context data.
     */
    readonly data: BehaviorSubjectLike<Context>
}

/** Create a {@link ContextService} instance. */
export function createContextService({
    clientApplication,
}: Pick<PlatformContext, 'clientApplication'>): ContextService {
    return {
        data: new BehaviorSubject<Context>({
            'clientApplication.isSourcegraph': clientApplication === 'sourcegraph',

            // Arbitrary, undocumented versioning for extensions that need different behavior for different
            // Sourcegraph versions.
            //
            // TODO: Make this more advanced if many extensions need this (although we should try to avoid
            // extensions needing this).
            'clientApplication.extensionAPIVersion.major': 3,
        }),
    }
}
