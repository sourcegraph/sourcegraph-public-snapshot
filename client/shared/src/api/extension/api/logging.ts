import { map } from 'rxjs/operators'
import sourcegraph from 'sourcegraph'

import { Settings } from '@sourcegraph/client-api'

import { ExtensionHostState } from '../extensionHostState'

/**
 * Sets active loggers extension host state based on user settings.
 *
 * @param state Extension host state
 */
export function setActiveLoggers(
    state: Pick<ExtensionHostState, 'settings' | 'activeLoggers'>
): sourcegraph.Unsubscribable {
    const activeLoggers = state.settings.pipe(map(settings => getActiveLoggersFromSettings(settings.final)))

    return activeLoggers.subscribe(activeLoggers => (state.activeLoggers = activeLoggers))
}

export function getActiveLoggersFromSettings(settings: Settings): Set<string> {
    if (!settings || !settings['extensions.activeLoggers']) {
        return new Set<string>()
    }

    const activeLoggers = settings['extensions.activeLoggers']

    if (!Array.isArray(activeLoggers)) {
        return new Set<string>()
    }

    return new Set<string>(activeLoggers)
}
