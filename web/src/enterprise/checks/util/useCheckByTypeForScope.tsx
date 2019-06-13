import { useEffect, useState } from 'react'
import { combineLatest, from, NEVER, of, Subscription } from 'rxjs'
import { catchError, delay, map, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { CheckID } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'

interface CheckData {
    id: CheckID
    provider: Pick<sourcegraph.CheckProvider, Exclude<keyof sourcegraph.CheckProvider, 'information'>>
    information: sourcegraph.CheckInformation
}

const LOADING: 'loading' = 'loading'

/**
 * Wait this long for the check provider to be registered (to allow time for the extension to
 * activate) before showing "not found".
 */
const CHECK_PROVIDER_REGISTRATION_DELAY = 5000 // ms

/**
 * A React hook that observes a single check (looked up by ID) for a particular scope.
 *
 * @param name The check name.
 * @param scope The scope in which to compute the check.
 */
export const useCheckByTypeForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    id: CheckID,
    scope: sourcegraph.CheckContext<any>['scope']
): typeof LOADING | CheckData | null | ErrorLike => {
    const [checkOrError, setCheckOrError] = useState<typeof LOADING | CheckData | null | ErrorLike>(LOADING)
    useEffect(() => {
        const idObj: CheckID = { type: id.type, id: id.id } // avoid useEffect rerunning for same type and id when object differs
        const subscriptions = new Subscription()
        subscriptions.add(
            combineLatest([
                extensionsController.services.checks.observeCheck(scope, idObj).pipe(
                    switchMap(provider =>
                        provider
                            ? from(provider.information).pipe(
                                  map(information => ({ id: idObj, provider, information }))
                              )
                            : of(null)
                    ),
                    startWith(LOADING)
                ),
                of(true).pipe(
                    delay(CHECK_PROVIDER_REGISTRATION_DELAY),
                    startWith(false)
                ),
            ])
                .pipe(
                    switchMap(([check, isDelayElapsed]) => (check || isDelayElapsed ? of(check) : NEVER)),
                    catchError(err => [asError(err)])
                )
                .subscribe(setCheckOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.checks, scope, id.type, id.id])
    return checkOrError
}
