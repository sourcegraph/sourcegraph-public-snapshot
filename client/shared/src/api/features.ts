import type { Remote } from 'comlink'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'

import { wrapRemoteObservable } from './client/api/common'
import type { FlatExtensionHostAPI } from './contract'

/**
 * Typically used to display loading indicators re: ready state of extension features
 */
export const haveInitialExtensionsLoaded = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(
            switchMap(extensionHost => wrapRemoteObservable(extensionHost.haveInitialExtensionsLoaded()))
        ),
    () => 'haveInitialExtensionsLoaded' // only one instance
)
