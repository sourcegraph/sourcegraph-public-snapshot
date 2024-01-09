import { FunctionComponent, useEffect, useState } from 'react'

import classNames from 'classnames'

import styles from './HubSpotForm.module.scss'

declare global {
    interface Window {
        hbspt?: {
            forms: {
                create: ({
                    region,
                    portalId,
                    formId,
                    target,
                    formInstanceId,
                    onFormSubmit,
                    onFormSubmitted,
                    onFormReady,
                }: HubSpotAPIProps) => CreateHubSpotFormProps
            }
        }
        jQuery: () => void
        ClearbitForHubspot: {
            forms: []
            addForm: () => void
        }
    }
}

interface HubSpotAPIProps {
    region?: string
    portalId: string
    formId: string
    target: string
    formInstanceId?: string
    onFormSubmit?: (object: { data: { name: string; value: string }[] }) => void
    onFormReady?: ($form: CreateHubSpotFormProps) => void
    onFormSubmitted?: () => void
    inlineMessage?: string
}

interface CreateHubSpotFormProps {
    [index: number]: HTMLFormElement
    formId: string
    onFormReady?: ($form: HTMLFormElement) => void
    onFormSubmitted?: () => void
    inlineMessage?: string
}

export interface HubSpotFormProps {
    formId?: string
    masterFormName?: 'qualificationSurvey'
    onFormSubmitted?: () => void
    onFormReady?: ($form: HTMLFormElement) => void
    inlineMessage?: string
    overrideFormShorten?: boolean
    userId?: string
    userEmail?: string
}

/**
 * These are our Master Forms that are used throughout our codebase. Each
 * masterFormName is used as an identifier to map to a specific master form id.
 */
const masterForms: { [key: string]: string } = {
    // Customer qualification survey
    qualificationSurvey: '85548efc-a879-4553-9ef0-a8da8fdcf541',
}

// HubSpot script integration
const hubSpotScript = 'https://js.hsforms.net/forms/v2.js'

// Third party script integrations
const jQueryScript = 'https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js'
const clearbitScript =
    '!function(e){var o=document.getElementsByTagName("script")[0];if("object"==typeof e.ClearbitForHubspot)return console.log("Clearbit For HubSpot included more than once"),!1;e.ClearbitForHubspot={},e.ClearbitForHubspot.forms=[],e.ClearbitForHubspot.addForm=function(o){var t=o[0];"function"==typeof e.ClearbitForHubspot.onFormReady?e.ClearbitForHubspot.onFormReady(t):e.ClearbitForHubspot.forms.push(t)};var t=document.createElement("script");t.async=!0,t.src="https://hubspot.clearbit.com/v1/forms/pk_a66b9ed76e62c713c06aab39bfae7234/forms.js",o.parentNode.insertBefore(t,o),e.addEventListener("message",function(o){if("hsFormCallback"===o.data.type&&"onFormReady"===o.data.eventName)if(document.querySelectorAll(\'form[data-form-id="\'+o.data.id+\'"]\').length>0)e.ClearbitForHubspot.addForm(document.querySelectorAll(\'form[data-form-id="\'+o.data.id+\'"]\'));else if(document.querySelectorAll("iframe.hs-form-iframe").length>0){document.querySelectorAll("iframe.hs-form-iframe").forEach(function(t){t.contentWindow.document.querySelectorAll(\'form[data-form-id="\'+o.data.id+\'"]\').length>0&&e.ClearbitForHubspot.addForm(t.contentWindow.document.querySelectorAll(\'form[data-form-id="\'+o.data.id+\'"]\'))})}})}(window);'

// Gets a script element by its id
const getScriptElement = (id: string): HTMLScriptElement | Element | null => document.querySelector(`#${id}`)

/**
 * This loads a script element and appends it to the document head
 *
 * @param id - a unique identifier for the script element
 * @param script - the script src (whether it's used for the script tag's src or innerHTML)
 * @param innerHTML - whether or not to assign the script to the script tag's src attribute or append to it's innerHTML
 * @returns - an HTML Script Element
 */
const loadScriptElement = (
    id: string,
    script: string,
    innerHTML?: boolean
): Promise<HTMLScriptElement | Element | null> =>
    new Promise(resolve => {
        const scriptElement = getScriptElement(id)

        if (!scriptElement) {
            const newScriptElement = document.createElement('script')
            newScriptElement.setAttribute('id', id)
            newScriptElement.setAttribute('async', '')
            if (innerHTML) {
                newScriptElement.innerHTML = script
            } else {
                newScriptElement.src = script
            }
            document.head.append(newScriptElement)
            resolve(newScriptElement)
        }

        resolve(scriptElement)
    })

/**
 * This loads all necessary scripts and third party integrations
 */
const loadAllScripts = async (): Promise<void> => {
    const loadingScripts: Promise<HTMLScriptElement | Element | null>[] = []

    if (!window.hbspt) {
        loadingScripts.push(loadScriptElement('hubspot', hubSpotScript))
    }
    if (!window.jQuery) {
        loadingScripts.push(loadScriptElement('jquery', jQueryScript))
    }
    if (!window.ClearbitForHubspot) {
        loadingScripts.push(loadScriptElement('clearbit', clearbitScript, true))
    }

    try {
        await Promise.all(loadingScripts)
    } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Error loading scripts:', error)
    }
}

/**
 * This creates the HubSpot form with the configuration options.
 * See: https://legacydocs.hubspot.com/docs/methods/forms/advanced_form_options
 *
 * @param CreateHubSpotFormProps - object props passed to createHubSpotForm
 * @param CreateHubSpotFormProps.formId - the form's id
 * @param CreateHubSpotFormProps.onFormReady - callback after form is built
 * @param CreateHubSpotFormProps.onFormSubmitted - callback after data is sent
 * @param CreateHubSpotFormProps.inlineMessage - form submission message
 */
function createHubSpotForm({ formId, onFormReady, onFormSubmitted, inlineMessage }: CreateHubSpotFormProps): void {
    const hbsptCreateForm = (): void => {
        window.hbspt?.forms.create({
            region: 'na1',
            portalId: '2762526',
            formId,
            target: '#form-target',
            onFormReady: (form: CreateHubSpotFormProps) => {
                if (onFormReady) {
                    onFormReady(form[0])
                }
            },
            onFormSubmitted,
            inlineMessage,
        })
    }

    if (window.hbspt) {
        hbsptCreateForm()

        return
    }

    // When the HubSpot script is loaded, create the form with the config
    getScriptElement('hubspot')?.addEventListener('load', hbsptCreateForm)
}

// This gets called when the HubSpot form is ready
const onHubsportFormReady = (form: HTMLFormElement, userId?: string, userEmail?: string): void => {
    /**
     * This allows you to populate hidden form fields with values
     *
     * @param formField - the form field name
     * @param value - the value to populate
     */
    const populateHiddenFormField = (formField: string, value: string): void => {
        const input = form.querySelector(`input[name="${formField}"]`) as HTMLInputElement
        if (input && !input.value) {
            input.value = value || ''
        }
    }

    /**
     * If the form is ready and visible in the DOM, gather all cookie and
     * session data and populate hidden form fields.
     */
    if (form) {
        const getAllCookies: { [index: string]: string } = document.cookie
            .split(';')
            .reduce((key, string) => Object.assign(key, { [string.split('=')[0].trim()]: string.split('=')[1] }), {})
        const { sourcegraphAnonymousUid, sourcegraphSourceUrl } = getAllCookies
        const landingSource: string = sessionStorage.getItem('landingSource') || ''
        const firstSourceURL: string =
            sourcegraphSourceUrl?.includes('redacted') || !sourcegraphSourceUrl ? landingSource : sourcegraphSourceUrl

        if (userId) {
            populateHiddenFormField('user_id', userId)
        } else {
            populateHiddenFormField('anonymous_user_id', sourcegraphAnonymousUid)
        }
        if (userEmail) {
            populateHiddenFormField('email', userEmail)
        }
        populateHiddenFormField('first_source_url', firstSourceURL)
        populateHiddenFormField('form_submission_source', window.location.href)
    }
}

/**
 * The HubSpot form component.
 *
 * @param options - option props
 * @param options.formId - an optional form id
 * @param options.onFormSubmitted - a callback that runs after a form submission
 * @param options.inlineMessage - a message to display after a form submission
 * @param options.userId - authenticated user's id
 * @param options.userEmail - authenticated user's email
 * @returns - a div element with an id where the HubSpot form renders
 */
export const HubSpotForm: FunctionComponent<HubSpotFormProps> = ({
    formId,
    masterFormName,
    onFormSubmitted,
    onFormReady,
    inlineMessage = 'Thank you for your feedback!',
    userEmail,
    userId,
}) => {
    const [formCreated, setFormCreated] = useState<boolean>(false)

    useEffect(() => {
        // Set the master form id if it's provided
        let masterFormId = ''
        if (masterFormName) {
            masterFormId = masterForms[masterFormName]
        }

        // Load all scripts
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        loadAllScripts()

        if (!formCreated) {
            createHubSpotForm({
                formId: formId || masterFormId,
                onFormReady: form => {
                    onFormReady?.(form)
                    onHubsportFormReady(form, userId, userEmail)
                },
                onFormSubmitted,
                inlineMessage,
            })

            setFormCreated(true)
        }
    }, [formId, onFormSubmitted, inlineMessage, formCreated, masterFormName, onFormReady, userId, userEmail])

    return <div id="form-target" data-testid="hubspot-form-container" className={classNames(styles.container)} />
}
