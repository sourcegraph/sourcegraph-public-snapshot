import { Observable, of, ReplaySubject, Subject } from 'rxjs'
import { catchError, map, switchMap, throttleTime } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'
import { ViewerSettingsResult, ViewerSettingsVariables } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ISettingsCascade } from '@sourcegraph/shared/src/schema'
import { EMPTY_SETTINGS_CASCADE, gqlToCascade, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

export function initializeSourcegraphSettings(
    requestGraphQL: PlatformContext['requestGraphQL']
): {
    settings: Observable<SettingsCascadeOrError>
    refreshSettings: () => void
    subscription: { dispose: () => void }
} {
    const settings = new ReplaySubject<SettingsCascadeOrError>(1)

    const refreshes = new Subject<void>()

    // Throttle refreshes for one hour.
    const ONE_HOUR_MS = 60 * 60 * 1000

    const subscription = refreshes
        .pipe(
            throttleTime(ONE_HOUR_MS, undefined, { leading: true, trailing: true }),
            switchMap(() =>
                requestGraphQL<ViewerSettingsResult, ViewerSettingsVariables>({
                    request: viewerSettingsQuery,
                    variables: {},
                    mightContainPrivateInfo: true,
                })
            ),
            map(({ data, errors }) => {
                if (!data?.viewerSettings) {
                    throw createAggregateError(errors)
                }
                return gqlToCascade(data?.viewerSettings as ISettingsCascade)
            }),
            catchError(() => {
                console.warn('Failed to load settings')
                return of(EMPTY_SETTINGS_CASCADE)
            })
        )
        .subscribe(settingsCascade => {
            settings.next(settingsCascade as SettingsCascadeOrError)
        })

    // Initial settings
    refreshes.next()

    return {
        settings: settings.asObservable(),
        refreshSettings: () => {
            refreshes.next()
        },
        subscription: { dispose: () => subscription.unsubscribe() },
    }
}
