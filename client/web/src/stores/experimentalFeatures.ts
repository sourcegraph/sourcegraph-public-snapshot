import create from 'zustand'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

const defaultSettings: SettingsExperimentalFeatures = {
    codeMonitoring: true,
    showEnterpriseHomePanels: true,
    /**
     * Whether we show the mulitiline editor at /search/console
     */
    showMultilineSearchConsole: false,
    showOnboardingTour: true,
    showSearchContext: true,
    showSearchContextManagement: true,
    showSearchNotebook: true,
    showComputeComponent: false,
    codeMonitoringWebHooks: false,
    showCodeMonitoringLogs: true,
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
