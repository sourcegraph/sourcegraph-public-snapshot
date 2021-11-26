import ApplicationBracketsOutlineIcon from 'mdi-react/ApplicationBracketsOutlineIcon'

import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

/**
 * Feature guard for catalog.
 *
 * @param settingsCascade - settings cascade object
 */
export function isCatalogEnabled(settingsCascade: SettingsCascadeOrError): boolean {
    if (isErrorLike(settingsCascade.final)) {
        return false
    }

    return Boolean(settingsCascade.final?.experimentalFeatures?.catalog)
}

export const CatalogIcon = ApplicationBracketsOutlineIcon

/**
 * Common props for components needing to render catalog components.
 */
export interface CatalogProps {
    catalogEnabled?: boolean
}
