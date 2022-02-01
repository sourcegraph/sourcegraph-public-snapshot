import React, { useState, useCallback, useEffect } from 'react'
import * as uuid from 'uuid'

const HUBSPOT_EXTERNAL_SCRIPT = '//js.hsforms.net/forms/v2.js'

let globalHubSpotScriptInsertedPromise: null | Promise<void> = null

interface WindowWithHubspot extends Window {
    readonly hbspt: {
        forms: {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            create(options: any): void
        }
    }
}

// All available options can be found in the HubSpot documentation.
//
// https://legacydocs.hubspot.com/docs/methods/forms/advanced_form_options
interface HubSpotConfig {
    // User's portal ID
    portalId: string
    // Unique ID of the form you wish to build
    formId: string
}

export function useHubSpotForm(config: HubSpotConfig): React.ReactNode {
    const [isScriptLoaded, setIsScriptLoaded] = useState<boolean>(false)
    const [isFormRendered, setIsFormRendered] = useState<boolean>(false)
    const [containerId] = useState(() => uuid.v4())

    const onLoad = useCallback(() => {
        setIsScriptLoaded(true)
    }, [])
    const onError = useCallback(() => {
        throw new Error('Failed to load HubSpot form')
    }, [])

    useEffect(() => {
        if (globalHubSpotScriptInsertedPromise !== null) {
            // If the script was already added by any other callers of the hook, we don't have to
            // load it again.
            globalHubSpotScriptInsertedPromise.then(onLoad, onError)
            return
        }
        globalHubSpotScriptInsertedPromise = new Promise((resolve, reject) => {
            const script = document.createElement('script')
            script.src = HUBSPOT_EXTERNAL_SCRIPT
            script.async = true
            script.addEventListener('load', () => resolve())
            script.addEventListener('error', () => reject(new Error('Error loading script tag')))
            document.body.append(script)
        })
        globalHubSpotScriptInsertedPromise.then(onLoad, onError)
    }, [onLoad, onError])

    useEffect(() => {
        if (isScriptLoaded && !isFormRendered) {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
            const windowWithHubspot: WindowWithHubspot = window as any
            windowWithHubspot.hbspt.forms.create({ ...config, target: `#hs-${containerId}` })
            setIsFormRendered(true)
        }
    }, [isScriptLoaded, isFormRendered, config, containerId])

    return <div id={`hs-${containerId}`} />
}
