import { Observable, ReplaySubject, Subject } from 'rxjs'
import { map, switchMap, throttleTime } from 'rxjs/operators'
import { Disposable } from 'vscode'

import { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'
import { ViewerSettingsResult } from '@sourcegraph/shared/src/graphql-operations'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { ISettingsCascade } from '@sourcegraph/shared/src/graphql/schema'
import { gqlToCascade, Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'

export function initializeSourcegraphSettings(
    requestGraphQL: (request: string, variables: any) => Promise<GraphQLResult<any>>,
    subscriptions: Disposable[]
): {
    settings: Observable<SettingsCascadeOrError<Settings>>
    refreshSettings: () => void
} {
    const settings = new ReplaySubject<SettingsCascadeOrError<Settings>>(1)

    const refreshes = new Subject<void>()

    // Throttle refreshes for one hour.
    const ONE_HOUR_MS = 60 * 60 * 1000

    const subscription = refreshes
        .pipe(
            throttleTime(ONE_HOUR_MS, undefined, { leading: true, trailing: true }),
            switchMap(() => requestGraphQL(viewerSettingsQuery, {})),
            map(({ data, errors }: GraphQLResult<ViewerSettingsResult>) => {
                if (!data?.viewerSettings) {
                    throw createAggregateError(errors)
                }

                return gqlToCascade(data?.viewerSettings as ISettingsCascade)
            })
        )
        .subscribe(settingsCascade => {
            settings.next(settingsCascade)
        })
    subscriptions.push({ dispose: () => subscription.unsubscribe() })

    // Initial settings
    refreshes.next()

    return {
        settings: settings.asObservable(),
        refreshSettings: () => {
            refreshes.next()
        },
    }
}
