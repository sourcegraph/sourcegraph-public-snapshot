import { SettingsCascade } from '../../settings/settings'
import { Remote, proxyMarker } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { ReplaySubject } from 'rxjs'
import { ExposedToClient, CalledFromExtHost } from '../contract'

// This holds the entire Ext Host state
// as a single plain object
export interface ExtState {
    settings?: Readonly<SettingsCascade<object>>
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    exposedToMain: ExposedToClient
}

/**
 * Holds internally ExtState and manages communication with the Client
 * Returns initialized public Ext API ready for consumption and API object marshaled into Client
 * NOTE that this function will slowly merge with the one in extensionHost.ts
 *
 * @param mainAPI
 */
export const initNewExtensionAPI = (mainAPI: Remote<CalledFromExtHost>): InitResult => {
    const state: ExtState = {}

    const configChanges = new ReplaySubject<void>(1)

    const exposedToMain: ExposedToClient = {
        [proxyMarker]: true,
        updateConfigurationData: data => {
            state.settings = Object.freeze(data)
            configChanges.next()
        },
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
            update: (key, value) => mainAPI.changeConfiguration({ path: [key as string | number], value }),
            toJSON: () => snapshot,
        }
        return configuration
    }

    return {
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        exposedToMain,
    }
}
