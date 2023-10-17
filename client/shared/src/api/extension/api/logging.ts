import type { Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'

import type { Settings } from '../../../settings/settings'
import type { ExtensionHostState } from '../extensionHostState'

/**
 * Sets active loggers extension host state based on user settings.
 *
 * @param state Extension host state
 */
export function setActiveLoggers(state: Pick<ExtensionHostState, 'settings' | 'activeLoggers'>): Unsubscribable {
    const activeLoggers = state.settings.pipe(map(settings => getActiveLoggersFromSettings(settings.final)))

    return activeLoggers.subscribe(activeLoggers => (state.activeLoggers = activeLoggers))
}

export function getActiveLoggersFromSettings(settings: Settings): Set<string> {
    if (!settings?.['extensions.activeLoggers']) {
        return new Set<string>()
    }

    const activeLoggers = settings['extensions.activeLoggers']

    if (!Array.isArray(activeLoggers)) {
        return new Set<string>()
    }

    return new Set<string>(activeLoggers)
}
