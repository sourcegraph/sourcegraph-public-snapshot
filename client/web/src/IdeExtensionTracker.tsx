import React, { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

const ONE_MONTH = 1000 * 60 * 60 * 24 * 30

/**
 * This component uses UTM parameters to detect incoming traffic from our IDE extensions (VS Code
 * and JetBrains) and updates a temporary setting whenever these are found.
 */
export const IdeExtensionTracker: React.FunctionComponent = () => {
    const location = useLocation()

    const [, setLastVSCodeDetection] = useTemporarySetting('integrations.vscode.lastDetectionTimestamp', 0)
    const [, setLastJetBrainsDetection] = useTemporarySetting('integrations.jetbrains.lastDetectionTimestamp', 0)

    // We only want to capture the IDE UTM parameters on the first page load. In order to avoid
    // rerunning the effect below whenever location change, we instead capture a reference that we
    // never update.
    const locationReference = useRef(location)

    useEffect(() => {
        const parameters = new URLSearchParams(locationReference.current.search)
        const utmProductName = parameters.get('utm_product_name')
        const utmMedium = parameters.get('utm_medium')
        const utmSource = parameters.get('utm_source')

        if (utmProductName === 'IntelliJ IDEA') {
            setLastJetBrainsDetection(Date.now())
        } else if (utmMedium === 'VSCIDE' || utmSource?.toLowerCase().startsWith('vscode')) {
            setLastVSCodeDetection(Date.now())
        }
    }, [setLastJetBrainsDetection, setLastVSCodeDetection])

    return null
}

export const useIsActiveIdeIntegrationUser = (): undefined | boolean => {
    const [lastVSCodeDetection] = useTemporarySetting('integrations.vscode.lastDetectionTimestamp', 0)
    const [lastJetBrainsDetection] = useTemporarySetting('integrations.jetbrains.lastDetectionTimestamp', 0)
    const [now] = useState<number>(Date.now())

    if (lastVSCodeDetection === undefined || lastJetBrainsDetection === undefined) {
        return undefined
    }

    if (now - lastVSCodeDetection < ONE_MONTH) {
        return true
    }
    if (now - lastJetBrainsDetection < ONE_MONTH) {
        return true
    }
    return false
}
