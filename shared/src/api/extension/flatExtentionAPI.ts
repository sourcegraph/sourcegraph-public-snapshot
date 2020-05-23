import { SettingsCascade } from '../../settings/settings'
import { Remote } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { ReplaySubject, Subject, BehaviorSubject } from 'rxjs'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'

/**
 * Holds the entire state exposed to the extension host
 * as a single plain object
 */
export interface ExtState {
    settings?: Readonly<SettingsCascade<object>>

    roots: readonly sourcegraph.WorkspaceRoot[]
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: PartialWorkspaceNamespace
    exposedToMain: FlatExtHostAPI
}

/**
 * mimics sourcegraph.workspace namespace without documents
 */
export interface PartialWorkspaceNamespace {
    roots: readonly sourcegraph.WorkspaceRoot[]
    onDidChangeRoots: sourcegraph.Subscribable<void>
    rootChanges: sourcegraph.Subscribable<void>
    versionContext: string | undefined
    versionContextChanges: sourcegraph.Subscribable<string | undefined>
}
/**
 * Holds internally ExtState and manages communication with the Client
 * Returns initialized public Ext API ready for consumption and API object marshaled into Client
 * NOTE that this function will slowly merge with the one in extensionHost.ts
 *
 * @param mainAPI
 */
export const initNewExtensionAPI = (mainAPI: Remote<MainThreadAPI>): InitResult => {
    const state: ExtState = { roots: [] }

    const configChanges = new ReplaySubject<void>(1)

    const rootChanges = new Subject<void>()
    // TODO (simon) this holds implicit state of the version context
    // move it to ExtState
    const versionContextChanges = new BehaviorSubject<string | undefined>(undefined)

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
        syncVersionContext: ctx => versionContextChanges.next(ctx),
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
        get roots() {
            return state.roots
        },
        get versionContext() {
            return versionContextChanges.value
        },
        onDidChangeRoots: rootChanges,
        rootChanges,
        versionContextChanges: versionContextChanges.asObservable(),
    }

    return {
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        exposedToMain,
        workspace,
    }
}
