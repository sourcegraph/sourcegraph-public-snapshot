import { Remote, proxyMarker } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { ExposedToClient, CalledFromExtHost } from '../contract'

// NOTE: this is for demo purposes at the moment.
// This is curently just inlined in connection.ts
export const initMainAPI = (
    ext: Remote<ExposedToClient>,
    cleanup: Subscription,
    ctx: Pick<PlatformContext, 'updateSettings' | 'settings'>
): CalledFromExtHost => {
    // Settings
    cleanup.add(
        from(ctx.settings)
            .pipe(
                switchMap(settings => {
                    if (isSettingsValid(settings)) {
                        return ext.updateConfigurationData(settings)
                    }
                    return []
                })
            )
            .subscribe()
    )

    const mainAPI: CalledFromExtHost = {
        [proxyMarker]: true,
        changeConfiguration: edit => updateSettings(ctx, edit),
    }

    return mainAPI
}
