import React, { useCallback, useMemo } from 'react'

import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { usePersistentCadence } from '../../hooks'
import { browserExtensionInstalled } from '../../tracking/analyticsUtils'
import { HOVER_COUNT_KEY, HOVER_THRESHOLD } from '../RepoContainer'

import { BrowserExtensionAlert } from './BrowserExtensionAlert'
import { NativeIntegrationAlert, NativeIntegrationAlertProps } from './NativeIntegrationAlert'

export interface ExtensionAlertProps {
    onExtensionAlertDismissed: () => void
}

interface InstallIntegrationsAlertProps
    extends Pick<NativeIntegrationAlertProps, 'externalURLs' | 'className'>,
        ExtensionAlertProps {
    codeHostIntegrationMessaging: 'native-integration' | 'browser-extension'
}

const CADENCE_KEY = 'InstallIntegrationsAlert.pageViews'
const DISPLAY_CADENCE = 5
const HAS_DISMISSED_ALERT_KEY = 'has-dismissed-extension-alert'

export const InstallIntegrationsAlert: React.FunctionComponent<InstallIntegrationsAlertProps> = ({
    codeHostIntegrationMessaging,
    externalURLs,
    className,
    onExtensionAlertDismissed,
}) => {
    const displayCTABasedOnCadence = usePersistentCadence(CADENCE_KEY, DISPLAY_CADENCE)
    const isBrowserExtensionInstalled = useObservable<boolean>(browserExtensionInstalled)
    const [hoverCount] = useLocalStorage<number>(HOVER_COUNT_KEY, 0)
    const [hasDismissedExtensionAlert, setHasDismissedExtensionAlert] = useLocalStorage<boolean>(
        HAS_DISMISSED_ALERT_KEY,
        false
    )
    const showExtensionAlert = useMemo(
        () =>
            isBrowserExtensionInstalled === false &&
            displayCTABasedOnCadence &&
            !hasDismissedExtensionAlert &&
            hoverCount >= HOVER_THRESHOLD,
        // Intentionally use useMemo() here without a dependency on hoverCount to only show the alert on the next reload,
        // to not cause an annoying layout shift from displaying the alert.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [hasDismissedExtensionAlert, isBrowserExtensionInstalled]
    )

    const onAlertDismissed = useCallback(() => {
        onExtensionAlertDismissed()
        setHasDismissedExtensionAlert(true)
    }, [onExtensionAlertDismissed, setHasDismissedExtensionAlert])

    if (!showExtensionAlert) {
        return null
    }

    if (codeHostIntegrationMessaging === 'native-integration') {
        return (
            <NativeIntegrationAlert
                className={className}
                onAlertDismissed={onAlertDismissed}
                externalURLs={externalURLs}
            />
        )
    }

    return <BrowserExtensionAlert className={className} onAlertDismissed={onAlertDismissed} />
}
