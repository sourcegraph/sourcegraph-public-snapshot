import { Remote } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { WorkspaceService } from './services/workspaceService'

export const initMainThreadAPI = (
    ext: Remote<FlatExtHostAPI>,
    platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'>,
    { roots, versionContext }: WorkspaceService
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

    // Workspace
    subscription.add(roots.subscribe(rs => {
        ext.syncRoots(rs || [])
    }))

    // eslint-disable-next-line no-void
    subscription.add(versionContext.subscribe(ctx => void ext.syncVersionContext(ctx)))

    const mainAPI: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
    }

    return [mainAPI, subscription]
}
