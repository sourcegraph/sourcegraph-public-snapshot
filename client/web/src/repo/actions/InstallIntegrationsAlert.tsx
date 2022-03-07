import React, { useCallback, useMemo } from 'react'

import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { usePersistentCadence } from '../../hooks'
import { useIsActiveIdeIntegrationUser } from '../../IdeExtensionTracker'
import { browserExtensionInstalled } from '../../tracking/analyticsUtils'
import { HOVER_COUNT_KEY, HOVER_THRESHOLD } from '../RepoContainer'

import { BrowserExtensionAlert } from './BrowserExtensionAlert'
import { IDEExtensionAlert } from './IdeExtensionAlert'
import { NativeIntegrationAlert, NativeIntegrationAlertProps } from './NativeIntegrationAlert'

export interface ExtensionAlertProps {
    onExtensionAlertDismissed: () => void
}

interface InstallIntegrationsAlertProps
    extends Pick<NativeIntegrationAlertProps, 'externalURLs' | 'className' | 'page'>,
        ExtensionAlertProps {
    codeHostIntegrationMessaging: 'native-integration' | 'browser-extension'
}

const CADENCE_KEY = 'InstallIntegrationsAlert.pageViews'
const DISPLAY_CADENCE = 6
const IDE_CTA_CADENCE_SHIFT = 3
export const HAS_DISMISSED_BROWSER_EXTENSION_ALERT_KEY = 'hasDismissedBrowserExtensionAlert'
export const HAS_DISMISSED_IDE_EXTENSION_ALERT_KEY = 'hasDismissedIdeExtensionAlert'

type CtaToDisplay = 'browser' | 'ide'

export const InstallIntegrationsAlert: React.FunctionComponent<InstallIntegrationsAlertProps> = ({
    codeHostIntegrationMessaging,
    externalURLs,
    className,
    page,
    onExtensionAlertDismissed,
}) => {
    const displayBrowserExtensionCTABasedOnCadence = usePersistentCadence(CADENCE_KEY, DISPLAY_CADENCE)
    const displayIDEExtensionCTABasedOnCadence = usePersistentCadence(
        CADENCE_KEY,
        DISPLAY_CADENCE,
        IDE_CTA_CADENCE_SHIFT
    )
    const isBrowserExtensionInstalled = useObservable<boolean>(browserExtensionInstalled)
    const isUsingIdeIntegration = useIsActiveIdeIntegrationUser()
    const [hoverCount] = useLocalStorage<number>(HOVER_COUNT_KEY, 0)
    const [hasDismissedBrowserExtensionAlert, setHasDismissedBrowserExtensionAlert] = useLocalStorage<boolean>(
        HAS_DISMISSED_BROWSER_EXTENSION_ALERT_KEY,
        false
    )
    const [hasDismissedIDEExtensionAlert, setHasDismissedIDEExtensionAlert] = useLocalStorage<boolean>(
        HAS_DISMISSED_IDE_EXTENSION_ALERT_KEY,
        false
    )

    const ctaToDisplay = useMemo<CtaToDisplay | undefined>(
        (): CtaToDisplay | undefined => {
            if (
                isBrowserExtensionInstalled === false &&
                displayBrowserExtensionCTABasedOnCadence &&
                !hasDismissedBrowserExtensionAlert &&
                hoverCount >= HOVER_THRESHOLD
            ) {
                return 'browser'
            }

            if (
                isUsingIdeIntegration === false &&
                displayIDEExtensionCTABasedOnCadence &&
                !hasDismissedIDEExtensionAlert
            ) {
                return 'ide'
            }

            return undefined
        },
        /**
         * Intentionally use useMemo() here without a dependency on hoverCount to only show the alert on the next reload,
         * to not cause an annoying layout shift from displaying the alert.
         */
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [
            displayBrowserExtensionCTABasedOnCadence,
            displayIDEExtensionCTABasedOnCadence,
            hasDismissedBrowserExtensionAlert,
            hasDismissedIDEExtensionAlert,
            isBrowserExtensionInstalled,
            isUsingIdeIntegration,
        ]
    )

    const onAlertDismissed = useCallback(() => {
        onExtensionAlertDismissed()
        if (ctaToDisplay === 'browser') {
            setHasDismissedBrowserExtensionAlert(true)
        }

        if (ctaToDisplay === 'ide') {
            setHasDismissedIDEExtensionAlert(true)
        }
    }, [
        ctaToDisplay,
        onExtensionAlertDismissed,
        setHasDismissedBrowserExtensionAlert,
        setHasDismissedIDEExtensionAlert,
    ])

    if (ctaToDisplay === 'browser') {
        if (codeHostIntegrationMessaging === 'native-integration') {
            return (
                <NativeIntegrationAlert
                    className={className}
                    page={page}
                    onAlertDismissed={onAlertDismissed}
                    externalURLs={externalURLs}
                />
            )
        }

        return <BrowserExtensionAlert className={className} page={page} onAlertDismissed={onAlertDismissed} />
    }

    if (ctaToDisplay === 'ide') {
        return <IDEExtensionAlert className={className} page={page} onAlertDismissed={onAlertDismissed} />
    }

    return null
}
