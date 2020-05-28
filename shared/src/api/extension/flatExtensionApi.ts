import { SettingsCascade } from '../../settings/settings'
import { Remote } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { ReplaySubject, Subject, Unsubscribable } from 'rxjs'
import { FlatExtHostAPI, MainThreadAPI, CommandHandle } from '../contract'

/**
 * Holds the entire state exposed to the extension host
 * as a single plain object
 */
export interface ExtState {
    settings?: Readonly<SettingsCascade<object>>

    // Workspace
    roots: readonly sourcegraph.WorkspaceRoot[]
    versionContext: string | undefined
    registeredCommands: Map<CommandHandle, (...args: any[]) => unknown>
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: PartialWorkspaceNamespace
    exposedToMain: FlatExtHostAPI
    // todo this is needed as a temp solution to getter problem
    state: Readonly<ExtState>
    commands: typeof sourcegraph['commands']
}

/**
 * mimics sourcegraph.workspace namespace without documents
 */
export type PartialWorkspaceNamespace = Omit<
    typeof sourcegraph['workspace'],
    'textDocuments' | 'onDidOpenTextDocument' | 'openedTextDocuments' | 'roots' | 'versionContext'
>
/**
 * Holds internally ExtState and manages communication with the Client
 * Returns initialized public Ext API ready for consumption and API object ready to be passed to the main thread
 * NOTE that this function will slowly merge with the one in extensionHost.ts
 *
 * @param mainAPI
 */
export const initNewExtensionAPI = (mainAPI: Remote<MainThreadAPI>): InitResult => {
    const state: ExtState = { roots: [], versionContext: undefined, registeredCommands: new Map() }

    const configChanges = new ReplaySubject<void>(1)

    const rootChanges = new Subject<void>()
    const versionContextChanges = new Subject<string | undefined>()

    const exposedToMain: FlatExtHostAPI = {
        // Configuration
        syncSettingsData: data => {
            state.settings = Object.freeze(data)
            configChanges.next()
        },

        // Workspace
        syncRoots: (roots): void => {
            state.roots = Object.freeze(roots.map(plain => ({ ...plain, uri: new URL(plain.uri) })))
            rootChanges.next()
        },
        syncVersionContext: ctx => {
            state.versionContext = ctx
            versionContextChanges.next(ctx)
        },

        // Commands
        // note that it is noop if there is an error or race condition with unregistering commands
        executeExtensionCommand: (handle, args) => state.registeredCommands.get(handle)?.(...args),
    }

    // Configuration
    const getConfiguration = <C extends object>(): sourcegraph.Configuration<C> => {
        if (!state.settings) {
            throw new Error('unexpected internal error: settings data is not yet available')
        }

        const snapshot = state.settings.final as Readonly<C>

        const configuration: sourcegraph.Configuration<C> & { toJSON: any } = {
            value: snapshot,
            get: key => snapshot[key],
            update: (key, value) => mainAPI.applySettingsEdit({ path: [key as string | number], value }),
            toJSON: () => snapshot,
        }
        return configuration
    }

    // Workspace
    const workspace: PartialWorkspaceNamespace = {
        onDidChangeRoots: rootChanges.asObservable(),
        rootChanges: rootChanges.asObservable(),
        versionContextChanges: versionContextChanges.asObservable(),
    }

    // Commands
    const commands: typeof sourcegraph['commands'] = {
        executeCommand: (cmd, args) => mainAPI.executeCommand(cmd, args),
        registerCommand: (cmd, callback) =>
            unsubPromise(mainAPI.registerCommand(cmd), {
                ok: handle => state.registeredCommands.set(handle, callback),
                unsub: handle => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    mainAPI.unregisterCommand(handle)
                    state.registeredCommands.delete(handle)
                },
            }),
    }

    return {
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        exposedToMain,
        workspace,
        state,
        commands,
    }
}

/**
 * a function that handles a promise and sync returns an Unsubscribable
 * This is needed to be able to do this:
 * ```
 * const {unsubscribe} = subscribePromise(promise, {
 *      ok: ()=> log('ok'),
 *      unsub: ()=>log('unsub')
 * })
 * unsubscribe()
 * ```
 * Note that in the example above log('ok') wont be called but log('unsub') will when the promise resolves
 *
 *
 * @param promise {@link Promise} that will result in a value
 * @param handlers for success and proper cleanup. Note: cleanup will be called even after unsubscribe()
 */
function unsubPromise<T>(
    promise: Promise<T>,
    {
        ok,
        unsub,
    }: {
        ok: (data: T) => void
        unsub: (data: T) => void
    }
): Unsubscribable {
    // state machine
    // keeps track of the current state of the promise execution
    let state: 'ok' | 'pending' | 'unsubbed' = 'pending'
    // valid only in ok state, would be nice to use descriminated union
    // but probably too much boilerplate to be worth it without a library
    let data: T | undefined

    promise
        .then(val => {
            if (state === 'pending') {
                state = 'ok'
                data = val
                ok(val)
            }
            if (state === 'unsubbed') {
                // even though we didn't call ok() we still might want to perform other cleanups
                unsub(val)
            }
        })
        .catch(e => {
            // Err?
        })

    return {
        unsubscribe: () => {
            const prev = state
            state = 'unsubbed'
            if (prev === 'ok') {
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                unsub(data!)
            }
        },
    }
}
