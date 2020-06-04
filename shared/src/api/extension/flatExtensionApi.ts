import { SettingsCascade } from '../../settings/settings'
import { Remote, proxy } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { BehaviorSubject, Subject } from 'rxjs'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { syncSubscription } from '../util'

/**
 * Holds the entire state exposed to the extension host
 * as a single plain object
 */
export interface ExtState {
    settings: Readonly<SettingsCascade<object>>

    // Workspace
    roots: readonly sourcegraph.WorkspaceRoot[]
    versionContext: string | undefined

    // Search
    queryTransformers: sourcegraph.QueryTransformer[]
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: PartialWorkspaceNamespace
    exposedToMain: FlatExtHostAPI
    // todo this is needed as a temp solution for getter problem
    state: Readonly<ExtState>
    commands: typeof sourcegraph['commands']
    search: typeof sourcegraph['search']
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
 * Returns the initialized public extension API pieces ready for consumption and the internal extension host API ready to be exposed to the main thread
 * NOTE that this function will slowly merge with the one in extensionHost.ts
 *
 * @param mainAPI
 */
export const initNewExtensionAPI = (
    mainAPI: Remote<MainThreadAPI>,
    initialSettings: Readonly<SettingsCascade<object>>
): InitResult => {
    const state: ExtState = { roots: [], versionContext: undefined, settings: initialSettings, queryTransformers: [] }

    const configChanges = new BehaviorSubject<void>(undefined)
    // Most extensions never call `configuration.get()` synchronously in `activate()` to get
    // the initial settings data, and instead only subscribe to configuration changes.
    // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.

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
        syncVersionContext: context => {
            state.versionContext = context
            versionContextChanges.next(context)
        },

        // Search
        transformSearchQuery: query =>
            // this is racy because a transformer can be executed after it unsubscribed
            state.queryTransformers.reduce(
                (queryPromise, transformer) =>
                    // transformer can be unsubscribed in the middle of transformation chain
                    // Note that we don't include new ones but exclude the ones that are unsubsribed
                    state.queryTransformers.includes(transformer)
                        ? queryPromise.then(q => transformer.transformQuery(q))
                        : queryPromise,
                Promise.resolve(query)
            ),
    }

    // Configuration
    const getConfiguration = <C extends object>(): sourcegraph.Configuration<C> => {
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
        executeCommand: (command, ...args) => mainAPI.executeCommand(command, args),
        registerCommand: (command, callback) => syncSubscription(mainAPI.registerCommand(command, proxy(callback))),
    }

    // Search
    // this is basically an optimization to skip round trip to the worker
    const notifyAboutQueryTransformerChanges = (): void => {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        mainAPI.notifyIfThereAreQueryTransformers(state.queryTransformers.length > 0)
    }

    const search: typeof sourcegraph['search'] = {
        registerQueryTransformer: transformer => {
            state.queryTransformers.push(transformer)
            notifyAboutQueryTransformerChanges()
            return {
                unsubscribe: () => {
                    notifyAboutQueryTransformerChanges()
                    state.queryTransformers = state.queryTransformers.filter(t => t !== transformer)
                },
            }
        },
    }

    return {
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        exposedToMain,
        workspace,
        state,
        commands,
        search,
    }
}
