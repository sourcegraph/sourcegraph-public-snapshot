import { Remote } from 'comlink'
import { updateSettings } from './services/settings'
import { Subscription, from, Unsubscribable } from 'rxjs'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'
import { switchMap } from 'rxjs/operators'
import { FlatExtHostAPI, MainThreadAPI, CommandHandle } from '../contract'
import { WorkspaceService } from './services/workspaceService'
import { CommandRegistry } from './services/command'

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
    let nextCmdHandle = 1
    const cmds = new Map<number, Unsubscribable>()
    const mainAPI: MainThreadAPI = {
        applySettingsEdit: edit => updateSettings(platformContext, edit),
        executeCommand: (command, args) => cmdRegistry.executeCommand({ command, arguments: args }),
        registerCommand: command => {
            const handle = nextCmdHandle
            const typedHandle = (handle as unknown) as CommandHandle
            nextCmdHandle += 1
            cmds.set(
                handle,
                cmdRegistry.registerCommand({
                    command,
                    run: args => ext.executeExtensionCommand(typedHandle, args),
                })
            )
            return typedHandle
        },
        unregisterCommand: handle => {
            const h = (handle as unknown) as number
            // eslint-disable-next-line no-unused-expressions
            cmds.get(h)?.unsubscribe()
            return cmds.delete(h)
        },
    }

    return [mainAPI, sub]
}
