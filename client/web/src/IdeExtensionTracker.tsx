import React, { useEffect, useRef, useState } from 'react'

const VSCODE_SETTING = 'integrations.vscode.lastDetectionTimestamp'
const JETBRAINS_SETTING = 'integrations.jetbrains.lastDetectionTimestamp'
const ONE_MONTH = 1000 * 60 * 60 * 24 * 30

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

/**
 * This component uses UTM parameters to detect incoming traffic from our IDE extensions (VS Code
 * and JetBrains) and updates a temporary setting whenever these are found.
 */
export const IdeExtensionTracker: React.FunctionComponent = () => {
    const [, setLastVSCodeDetection] = useTemporarySetting(VSCODE_SETTING, 0)
    const [, setLastJetBrainsDetection] = useTemporarySetting(JETBRAINS_SETTING, 0)

    // We only want to run the below effect once. Since the setter function changes over time, we
    // capture a copy in a ref to avoid passing it into the dependency array of the effect.
    const setLastVSCodeDetectionReference = useRef(setLastVSCodeDetection)
    const setLastJetBrainsDetectionReference = useRef(setLastJetBrainsDetection)

    useEffect(() => {
        const parameters = new URLSearchParams(location.search)
        const utmProductName = parameters.get('utm_product_name')
        const utmMedium = parameters.get('utm_medium')
        const utmSource = parameters.get('utm_source')

        if (utmProductName === 'IntelliJ IDEA') {
            setLastJetBrainsDetectionReference.current(Date.now())
        } else if (utmMedium === 'VSCIDE' || utmSource?.toLowerCase().startsWith('vscode')) {
            setLastVSCodeDetectionReference.current(Date.now())
        }
    }, [])

    return null
}

export const useIsUsingIdeIntegration = (): undefined | boolean => {
    const [lastVSCodeDetection] = useTemporarySetting(VSCODE_SETTING, 0)
    const [lastJetBrainsDetection] = useTemporarySetting(JETBRAINS_SETTING, 0)
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
