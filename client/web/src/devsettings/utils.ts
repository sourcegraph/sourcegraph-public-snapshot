import { getTemporarySettingOverride } from '@sourcegraph/shared/src/settings/temporary/localOverride'
import { TEMPORARY_SETTINGS_KEYS } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'

import { FEATURE_FLAGS } from '../featureFlags/featureFlags'
import { getFeatureFlagOverrideValue } from '../featureFlags/lib/feature-flag-local-overrides'

export function countOverrides(): { featureFlags: number; temporarySettings: number } {
    return {
        featureFlags: FEATURE_FLAGS.reduce(
            (sum, name) => sum + (getFeatureFlagOverrideValue(name) === null ? 0 : 1),
            0
        ),
        temporarySettings: TEMPORARY_SETTINGS_KEYS.reduce(
            (sum, name) => sum + (getTemporarySettingOverride(name) === null ? 0 : 1),
            0
        ),
    }
}
