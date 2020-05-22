import { Remote } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'

export const initMainThreadAPI = (
    ext: Remote<FlatExtHostAPI>,
    platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'>
): [MainThreadAPI, Subscription] => {
    const subscription = new Subscription()
    // Settings
    subscription.add(
        from(platformContext.settings)
            .pipe(
                switchMap(settings => {
                    if (isSettingsValid(settings)) {
                        return ext.syncSettingsData(settings)
                    }
                    return []
                })
            )
            .subscribe()
    )

    const mainAPI: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
    }

    return [mainAPI, subscription]
}
