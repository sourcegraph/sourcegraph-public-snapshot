import { Remote, proxy } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { WorkspaceService } from './services/workspaceService'
import { CommandRegistry } from './services/command'
import { ProxySubscription } from './api/common'

export const initMainThreadAPI = (
    ext: Remote<FlatExtHostAPI>,
    platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'>,
    { roots, versionContext }: WorkspaceService,
    cmdRegistry: Pick<CommandRegistry, 'executeCommand' | 'registerCommand'>
): [MainThreadAPI, Subscription] => {
    const sub = new Subscription()
    // Settings
    sub.add(
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
    sub.add(
        roots.subscribe(rs => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            ext.syncRoots(rs || [])
        })
    )
    sub.add(
        versionContext.subscribe(ctx => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            ext.syncVersionContext(ctx)
        })
    )

    // Commands
    const mainAPI: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
        executeCommand: (command, args) => cmdRegistry.executeCommand({ command, arguments: args }),
        registerCommand: (command, run) => {
            const subscription = new Subscription()
            subscription.add(cmdRegistry.registerCommand({ command, run }))
            subscription.add(new ProxySubscription(run))
            return proxy(subscription)
        },
    }

    return [mainAPI, sub]
}
