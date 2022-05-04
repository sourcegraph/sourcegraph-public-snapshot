import React, { useCallback, useMemo } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useLocalStorage } from '@sourcegraph/wildcard'

import { usePersistentCadence } from '../../hooks'
import { useIsActiveIdeIntegrationUser } from '../../IdeExtensionTracker'
import { useTourQueryParameters } from '../../tour/components/Tour/TourAgent'
import { useIsBrowserExtensionActiveUser } from '../../tracking/BrowserExtensionTracker'
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

type CtaToDisplay = 'browser' | 'ide'

export const InstallIntegrationsAlert: React.FunctionComponent<
    React.PropsWithChildren<InstallIntegrationsAlertProps>
> = ({ codeHostIntegrationMessaging, externalURLs, className, page, onExtensionAlertDismissed }) => {
    const displayBrowserExtensionCTABasedOnCadence = usePersistentCadence(CADENCE_KEY, DISPLAY_CADENCE)
    const displayIDEExtensionCTABasedOnCadence = usePersistentCadence(
        CADENCE_KEY,
        DISPLAY_CADENCE,
        IDE_CTA_CADENCE_SHIFT
    )
    const isBrowserExtensionActiveUser = useIsBrowserExtensionActiveUser()
    const isUsingIdeIntegration = useIsActiveIdeIntegrationUser()
    const [hoverCount] = useLocalStorage<number>(HOVER_COUNT_KEY, 0)
    const [hasDismissedBrowserExtensionAlert, setHasDismissedBrowserExtensionAlert] = useTemporarySetting(
        'cta.browserExtensionAlertDismissed',
        false
    )
    const [hasDismissedIDEExtensionAlert, setHasDismissedIDEExtensionAlert] = useTemporarySetting(
        'cta.ideExtensionAlertDismissed',
        false
    )

    const tourQueryParameters = useTourQueryParameters()

    const ctaToDisplay = useMemo<CtaToDisplay | undefined>(
        (): CtaToDisplay | undefined => {
            if (tourQueryParameters?.isTour) {
                return
            }

            if (
                isBrowserExtensionActiveUser === false &&
                displayBrowserExtensionCTABasedOnCadence &&
                hasDismissedBrowserExtensionAlert === false &&
                hoverCount >= HOVER_THRESHOLD
            ) {
                return 'browser'
            }

            if (
                isUsingIdeIntegration === false &&
                displayIDEExtensionCTABasedOnCadence &&
                hasDismissedIDEExtensionAlert === false
            ) {
                return 'ide'
            }

            return
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
            isBrowserExtensionActiveUser,
            isUsingIdeIntegration,
            tourQueryParameters,
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
