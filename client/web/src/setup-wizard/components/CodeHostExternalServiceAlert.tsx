import type { FC } from 'react'

import {
    ExternalServiceEditingDisabledAlert,
    ExternalServiceEditingTemporaryAlert,
} from '../../components/externalServices'

export const CodeHostExternalServiceAlert: FC = () => {
    const { extsvcConfigFileExists, extsvcConfigAllowEdits } = window.context

    const isEditingDisabled = extsvcConfigFileExists && !extsvcConfigAllowEdits
    const isEditingStateless = extsvcConfigFileExists && extsvcConfigAllowEdits

    if (isEditingDisabled) {
        return <ExternalServiceEditingDisabledAlert className="mb-2" />
    }

    if (isEditingStateless) {
        return <ExternalServiceEditingTemporaryAlert className="mb-2" />
    }

    // If nothing is specified that means everything is available manually
    // in site admin or setup wizard UI
    return null
}
