import create from 'zustand'

import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { SettingsExperimentalFeatures } from '../schema/settings.schema'

export const useExperimentalFeatures = create<SettingsExperimentalFeatures>(() => ({}))

export function setExperimentalFeaturesFromSettings(settingsCascade: SettingsCascadeOrError): void {
    const experimentalFeatures: SettingsExperimentalFeatures =
        (settingsCascade.final && !isErrorLike(settingsCascade.final) && settingsCascade.final.experimentalFeatures) ||
        {}

    useExperimentalFeatures.setState(experimentalFeatures, true)
}
