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

interface Props {
    hubSpotConfig: HubSpotConfig
    initialFormValues?: { [key: string]: string }
    onFormSubmitted?: () => void
    onError: (error: Error) => void
}

export function useHubSpotForm({ hubSpotConfig, onFormSubmitted, onError, initialFormValues }: Props): React.ReactNode {
    const [isScriptLoaded, setIsScriptLoaded] = useState<boolean>(false)
    const [isFormRendered, setIsFormRendered] = useState<boolean>(false)
    const [containerId] = useState(() => uuid.v4())

    const onLoad = useCallback(() => {
        setIsScriptLoaded(true)
    }, [])

    const { formId } = hubSpotConfig

    const onFormSubmittedCallback = useCallback(() => {
        if (typeof onFormSubmitted !== 'function') {
            return
        }
        onFormSubmitted()
    }, [onFormSubmitted])

    const onFormReadyCallback = useCallback(() => {
        if (!initialFormValues) {
            return
        }
        // Prefilling form fields following the examples from
        // https://legacydocs.hubspot.com/docs/methods/forms/advanced_form_options
        const iframeDocument = getFormDocument(`hs-${containerId}`)
        if (!iframeDocument) {
            return
        }
        for (const [field, value] of Object.entries(initialFormValues)) {
            const input = iframeDocument.querySelector<HTMLInputElement>(`input[name="${field}"]`)
            if (input) {
                input.value = value
                input.dispatchEvent(new Event('input', { bubbles: true }))
            }
        }
    }, [containerId, initialFormValues])

    const onMessage = useCallback(
        event => {
            if (
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                event.data.type === 'hsFormCallback' &&
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                event.data.id === formId
            ) {
                // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access,@typescript-eslint/no-unsafe-assignment
                const eventName = event.data.eventName

                if (eventName === 'onFormSubmit') {
                    onFormSubmittedCallback()
                } else if (eventName === 'onFormReady') {
                    onFormReadyCallback()
                }
            }
        },
        [formId, onFormSubmittedCallback, onFormReadyCallback]
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
            if (typeof onFormSubmitted === 'function') {
                window.addEventListener('message', onMessage)
                cleanup = () => window.removeEventListener
            }

            if (!isFormRendered) {
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
                const windowWithHubspot: WindowWithHubspot = window as any
                windowWithHubspot.hbspt.forms.create({ ...hubSpotConfig, target: `#hs-${containerId}` })
                setIsFormRendered(true)
            }
        }

        return cleanup
    }, [isScriptLoaded, isFormRendered, hubSpotConfig, containerId, onMessage, onFormSubmitted])

    return <div id={`hs-${containerId}`} />
}

function getFormDocument(containerId: string): undefined | Document {
    return document.querySelector<HTMLIFrameElement>(`#${containerId} > iframe`)?.contentDocument ?? undefined
}
