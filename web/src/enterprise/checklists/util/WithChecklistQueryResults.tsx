import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { Checklist } from '../checklist'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a checklist.
 */
export const useChecklist = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    scope: sourcegraph.ChecklistScope | sourcegraph.WorkspaceRoot
): typeof LOADING | Checklist | ErrorLike => {
    const [checklistOrError, setChecklistOrError] = useState<typeof LOADING | Checklist | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            extensionsController.services.checklist
                .observeChecklistItems(scope)
                .pipe(
                    map(checklistItems => ({ items: checklistItems })),
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setChecklistOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.checklist, scope])
    return checklistOrError
}
