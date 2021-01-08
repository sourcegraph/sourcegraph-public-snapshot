import { Remote } from 'comlink'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { memoizeObservable } from '../util/memoizeObservable'
import { wrapRemoteObservable } from './client/api/common'
import { FlatExtensionHostAPI } from './contract'

/**
 * Typically used to display loading indicators re: ready state of extension features
 */
export const haveInitialExtensionsLoaded = memoizeObservable(
    (extHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extHostAPI).pipe(
            switchMap(extensionHost => wrapRemoteObservable(extensionHost.haveInitialExtensionsLoaded()))
        ),
    () => 'haveInitialExtensionsLoaded' // only one instance
)
