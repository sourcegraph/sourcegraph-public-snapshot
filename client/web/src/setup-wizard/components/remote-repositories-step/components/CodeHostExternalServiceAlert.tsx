import { FC } from 'react'

import {
    ExternalServiceEditingDisabledAlert,
    ExternalServiceEditingTemporaryAlert,
} from '../../../../components/externalServices'

export const CodeHostExternalServiceAlert: FC = () => {
    const { extsvcConfigFileExists, extsvcConfigAllowEdits } = window.context

    const isEditingDisabled = extsvcConfigFileExists && !extsvcConfigAllowEdits
    const isEditingStateless = extsvcConfigFileExists && extsvcConfigAllowEdits

    if (isEditingDisabled) {
        return <ExternalServiceEditingDisabledAlert />
    }

    if (isEditingStateless) {
        return <ExternalServiceEditingTemporaryAlert />
    }

    // If nothing is specified that means everything is available manually
    // in site admin or setup wizard UI
    return null
}
