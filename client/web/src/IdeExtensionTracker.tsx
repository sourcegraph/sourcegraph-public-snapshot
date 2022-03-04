import React, { useEffect, useRef, useState } from 'react'

const VSCODE_SETTING = 'integrations.vscode.lastDetectionTimestamp'
const JETBRAINS_SETTING = 'integrations.jetbrains.lastDetectionTimestamp'
const ONE_MONTH = 1000 * 60 * 60 * 24 * 30

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

// @todo: Add explaination
// @todo: Add tests
// https://sourcegraph.test:3443/-/editor?remote_url=git%40github.com%3Asourcegraph%2Fsourcegraph-jetbrains.git&branch=main&file=src%2Fmain%2Fjava%2FOpenRevisionAction.java&editor=JetBrains&version=v1.2.2&start_row=68&start_col=26&end_row=68&end_col=26&utm_product_name=IntelliJ+IDEA&utm_product_version=2021.3.2
// https://sourcegraph.test:3443/-/editor?remote_url=git@github.com:sourcegraph/sourcegraph.git&branch=ps/detect-ide-extensions&file=client/web/src/tracking/util.ts&editor=VSCode&version=2.0.9&start_row=13&start_col=22&end_row=13&end_col=22&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=vsce-commands
// https://sourcegraph.test:3443/sign-up?editor=vscode&utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up
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

export const useIsUsingIDEIntegration = (): undefined | boolean => {
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
