import { FC } from 'react'

import {
    ExternalServiceEditingDisabledAlert,
    ExternalServiceEditingTemporaryAlert,
    ExternalServiceEditingAppLimitInPlaceAlert,
    ExternalServiceEditingAppLimitReachedAlert,
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

export const CodeHostRepositoriesAppLimitAlert: FC<{ className?: string }> = props => {
    const { sourcegraphAppMode } = window.context

    if (sourcegraphAppMode) {
        return <ExternalServiceEditingAppLimitInPlaceAlert className={props.className} />
    }

    return null
}

export const CodeHostAppLimit: FC<{ className?: string }> = props => {
    const { sourcegraphAppMode } = window.context

    if (sourcegraphAppMode) {
        return <ExternalServiceEditingAppLimitReachedAlert className={props.className} />
    }

    return null
}
