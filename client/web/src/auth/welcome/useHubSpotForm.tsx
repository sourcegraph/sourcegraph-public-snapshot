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
    // Callback the data is actually sent. This allows you to perform an action when the submission is fully complete,
    // such as displaying a confirmation or thank you message.
    onFormSubmitted?: () => void
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

    const { onFormSubmitted, formId } = config
    const onFormSubmittedCallback = useCallback(
        event => {
            if (typeof onFormSubmitted !== 'function') {
                return
            }

            if (
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                event.data.type === 'hsFormCallback' &&
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                event.data.eventName === 'onFormSubmit' &&
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                event.data.id === formId
            ) {
                onFormSubmitted()
            }
        },
        [onFormSubmitted, formId]
    )

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
        let cleanup = (): void => {}
        if (isScriptLoaded) {
            // The HubSpot API has an option to add form submission handlers. They do, however
            // require jQuery to be loaded in the site. We don't want do load jQuery just for this
            // functionality though.
            // As a workaround we rely on HubSpots global event API as descirbed here:
            //   https://legacydocs.hubspot.com/global-form-events
            // More context can be found here:
            //   https://github.com/escaladesports/react-hubspot-form/issues/22
            const { onFormSubmitted, ...rest } = config
            if (typeof onFormSubmitted === 'function') {
                window.addEventListener('message', onFormSubmittedCallback)
                cleanup = () => window.removeEventListener
            }

            if (!isFormRendered) {
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
                const windowWithHubspot: WindowWithHubspot = window as any
                windowWithHubspot.hbspt.forms.create({ ...rest, target: `#hs-${containerId}` })
                setIsFormRendered(true)
            }
        }

        return cleanup
    }, [isScriptLoaded, isFormRendered, config, containerId, onFormSubmittedCallback])

    return <div id={`hs-${containerId}`} />
}
