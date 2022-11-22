import { Extension } from '@codemirror/state'

import { isErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

export const enableTokenSelection = (settingsCascade: SettingsCascadeOrError<Settings>): boolean =>
    (settingsCascade.final &&
        !isErrorLike(settingsCascade.final) &&
        (settingsCascade.final['codeIntel.blobKeyboardNavigation'] === 'token' ||
            settingsCascade.final['experimentalFeatures.tokenSelection'] === true)) ??
    false

export function tokenSelection(): Extension {
    return []
}
