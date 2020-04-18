import { useMemo } from 'react'
import { Observable, combineLatest } from 'rxjs'
import { Contributions, ViewContribution, FormContribution, Evaluated } from '../../../../shared/src/api/protocol'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { map } from 'rxjs/operators'
import { ViewService, View } from '../../../../shared/src/api/client/services/viewService'

interface ViewData {
    /**
     * The view's contribution.
     */
    contribution: ViewContribution

    /**
     * The form, `undefined` if none is defined, or `null` if a form is defined but not found.
     */
    form?: FormContribution | null

    /**
     * The view's content.
     */
    view: View | null
}

/**
 * A React hook that returns the view and its associated contents (such as its form, if any),
 * `undefined` if loading, and `null` if not found.
 */
export const useView = (
    viewID: string,
    params: { [key: string]: string },
    contributions: Observable<Evaluated<Contributions>>,
    viewService: ViewService
): ViewData | undefined | null =>
    useObservable(
        useMemo(
            () =>
                combineLatest([
                    contributions.pipe(
                        map(contributions => {
                            const contribution = contributions.views?.find(({ id }) => id === viewID)
                            if (!contribution) {
                                return null
                            }
                            if (!contribution.form) {
                                return { contribution }
                            }
                            const form = contributions.forms?.find(({ id }) => id === contribution.form) ?? null
                            return { contribution, form }
                        })
                    ),
                    viewService.get(viewID, params),
                ]).pipe(map(([viewAndForm, view]) => (viewAndForm ? { ...viewAndForm, view } : null))),
            [contributions, viewID, viewService, params]
        )
    )
