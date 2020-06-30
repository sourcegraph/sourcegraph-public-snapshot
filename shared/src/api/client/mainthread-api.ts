import { Remote, proxy } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from, of } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap, concatMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { ProxySubscription } from './api/common'
import { Services } from './services'
import { proxySubscribable } from '../extension/api/common'

// for now it will partially mimic Services object but hopefully will be incrementally reworked in the process
export type MainThreadAPIDependencies = Pick<Services, 'commands' | 'workspace'>

export const initMainThreadAPI = (
    extensionHost: Remote<FlatExtHostAPI>,
    platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'>,
    dependencies: MainThreadAPIDependencies
): { api: MainThreadAPI; subscription: Subscription } => {
    const {
        workspace: { roots, versionContext },
        commands,
    } = dependencies

    const subscription = new Subscription()
    // Settings
    subscription.add(
        from(platformContext.settings)
            .pipe(
                switchMap(settings => {
                    if (isSettingsValid(settings)) {
                        return extensionHost.syncSettingsData(settings)
                    }
                    return []
                })
            )
            .subscribe()
    )

    // Workspace
    subscription.add(
        from(roots)
            .pipe(concatMap(roots => extensionHost.syncRoots(roots)))
            .subscribe()
    )
    subscription.add(
        from(versionContext)
            .pipe(concatMap(context => extensionHost.syncVersionContext(context)))
            .subscribe()
    )

    // Commands
    const api: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
        executeCommand: (command, args) => commands.executeCommand({ command, arguments: args }),
        registerCommand: (command, run) => {
            const subscription = new Subscription()
            subscription.add(commands.registerCommand({ command, run }))
            subscription.add(new ProxySubscription(run))
            return proxy(subscription)
        },
        getActiveExtensions: () => proxySubscribable(of([])),
        getScriptURLForExtension: url => Promise.resolve(url),
    }

    return { api, subscription }
}
