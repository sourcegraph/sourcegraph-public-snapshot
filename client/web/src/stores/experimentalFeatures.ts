import create from 'zustand'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

const defaultSettings: SettingsExperimentalFeatures = {
    codeMonitoring: true,
    /**
     * Whether we show the multiline editor at /search/console
     */
    showMultilineSearchConsole: false,
    showSearchContext: true,
    showSearchNotebook: true,
    codeMonitoringWebHooks: true,
    showCodeMonitoringLogs: true,
    codeInsightsCompute: false,
    editor: 'codemirror6',
    codeInsightsRepoUI: 'search-query-or-strict-list',
    applySearchQuerySuggestionOnEnter: false,
    setupWizard: false,
}

export const useExperimentalFeatures = create<SettingsExperimentalFeatures>(() => ({}))

export function setExperimentalFeaturesFromSettings(settingsCascade: SettingsCascadeOrError): void {
    const experimentalFeatures: SettingsExperimentalFeatures =
        (settingsCascade.final && !isErrorLike(settingsCascade.final) && settingsCascade.final.experimentalFeatures) ||
        {}

    useExperimentalFeatures.setState({ ...defaultSettings, ...experimentalFeatures }, true)
}

// For testing purposes only. Initializes the feature flags with the default values.
export function setExperimentalFeaturesForTesting(): void {
    useExperimentalFeatures.setState(defaultSettings, true)
}

/**
 * This is a helper function to hide the fact that experimental feature flags
 * are backed by a Zustand store
 */
export function getExperimentalFeatures(): SettingsExperimentalFeatures {
    return useExperimentalFeatures.getState()
}
