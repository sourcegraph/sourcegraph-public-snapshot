import { Remote, proxy } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap, concatMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { ProxySubscription } from './api/common'
import { Services } from './services'
import { TransformQuerySignature } from './services/queryTransformer'

// for now it will partially mimic Services object but hopefully will be incrementally reworked in the process
export type MainThreadAPIDependencies = Pick<Services, 'commands' | 'workspace'>

export const initMainThreadAPI = (
    extentionHost: Remote<FlatExtHostAPI>,
    platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'>,
    dependencies: MainThreadAPIDependencies
): { api: MainThreadAPI; subscription: Subscription; transformQuery: TransformQuerySignature } => {
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
                        return extentionHost.syncSettingsData(settings)
                    }
                    return []
                })
            )
            .subscribe()
    )

    // Workspace
    subscription.add(
        from(roots)
            .pipe(concatMap(roots => extentionHost.syncRoots(roots)))
            .subscribe()
    )
    subscription.add(
        from(versionContext)
            .pipe(concatMap(context => extentionHost.syncVersionContext(context)))
            .subscribe()
    )

    // Search
    // this is basically an optimization to skip round trip to the worker if we don't have any transformers
    // which is probably the norm
    let hasRegisteredTransformers = false
    const transformQuery: TransformQuerySignature = query =>
        hasRegisteredTransformers ? extentionHost.transformSearchQuery(query) : Promise.resolve(query)

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
        notifyIfThereAreQueryTransformers: yesThereIsSome => (hasRegisteredTransformers = yesThereIsSome),
    }

    return { api, subscription, transformQuery }
}
