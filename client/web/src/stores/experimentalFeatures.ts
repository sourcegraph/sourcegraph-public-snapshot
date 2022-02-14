import create from 'zustand'

import { SettingsCascadeOrError } from '@sourcegraph/client-api'
import { isErrorLike } from '@sourcegraph/common'
import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'

const defaultSettings: SettingsExperimentalFeatures = {
    apiDocs: true,
    codeMonitoring: true,
    showEnterpriseHomePanels: true,
    /**
     * Whether we show the mulitiline editor at /search/console
     */
    showMultilineSearchConsole: false,
    showOnboardingTour: true,
    showSearchContext: true,
    showSearchContextManagement: true,
    showSearchNotebook: false,
    codeMonitoringWebHooks: false,
}

export const useExperimentalFeatures = create<SettingsExperimentalFeatures>(() => ({}))

export function setExperimentalFeaturesFromSettings(settingsCascade: SettingsCascadeOrError): void {
    const experimentalFeatures: SettingsExperimentalFeatures =
        (settingsCascade.final && !isErrorLike(settingsCascade.final) && settingsCascade.final.experimentalFeatures) ||
        {}

    useExperimentalFeatures.setState({ ...defaultSettings, ...experimentalFeatures }, true)
}

/**
 * This is a helper function to hide the fact that experimental feature flags
 * are backed by a Zustand store
 */
export function getExperimentalFeatures(): SettingsExperimentalFeatures {
    return useExperimentalFeatures.getState()
}
