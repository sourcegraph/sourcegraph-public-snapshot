import '../../shared/polyfills'

import { fromEvent, Subscription } from 'rxjs'
import { first } from 'rxjs/operators'

import { setLinkComponent, AnchorLink } from '@sourcegraph/shared/src/components/Link'

import { determineCodeHost } from '../../shared/code-hosts/shared/codeHost'
import { injectCodeIntelligence } from '../../shared/code-hosts/shared/inject'
import { logger } from '../../shared/code-hosts/shared/util/logger'
import {
    EXTENSION_MARKER_ID,
    injectExtensionMarker,
    NATIVE_INTEGRATION_ACTIVATED,
} from '../../shared/code-hosts/sourcegraph/inject'
import { initSentry } from '../../shared/sentry'
import { CLOUD_SOURCEGRAPH_URL, getAssetsURL } from '../../shared/util/context'
import { featureFlags } from '../../shared/util/featureFlags'
import { assertEnvironment } from '../environmentAssertion'

const subscriptions = new Subscription()
window.addEventListener('unload', () => subscriptions.unsubscribe(), { once: true })

assertEnvironment('CONTENT')

const codeHost = determineCodeHost()
initSentry('content', codeHost?.type)

setLinkComponent(AnchorLink)

// Add style sheet and wait for it to load to avoid rendering unstyled elements (which causes an
// annoying flash/jitter when the stylesheet loads shortly thereafter).
function loadStyleSheet(options: { id: string; path: string }): HTMLLinkElement {
    const { id, path } = options

    let styleSheet = document.querySelector<HTMLLinkElement>(`#${id}`)
    // If does not exist, create
    if (!styleSheet) {
        styleSheet = document.createElement('link')
        styleSheet.id = id
        styleSheet.rel = 'stylesheet'
        styleSheet.type = 'text/css'
        styleSheet.href = browser.extension.getURL(path)
    }
    return styleSheet
}

// If stylesheet is not loaded yet, wait for it to load.
async function waitForStyleSheet(styleSheet: HTMLLinkElement): Promise<void> {
    if (!styleSheet.sheet) {
        await new Promise(resolve => {
            styleSheet.addEventListener('load', resolve, { once: true })
            // If not appended yet, append to <head>
            if (!styleSheet.parentNode) {
                document.head.append(styleSheet)
            }
        })
    }
}

/**
 * Main entry point into browser extension.
 */
async function main(): Promise<void> {
    logger.info('Browser extension is running')

    // Make sure DOM is fully loaded
    if (document.readyState !== 'complete' && document.readyState !== 'interactive') {
        await new Promise<Event>(resolve => document.addEventListener('DOMContentLoaded', resolve, { once: true }))
    }

    // Allow users to set this via the console.
    ;(window as any).sourcegraphFeatureFlags = featureFlags

    // Check if a native integration is already running on the page,
    // and abort execution if it's the case.
    // If the native integration was activated before the content script, we can
    // synchronously check for the presence of the extension marker.
    if (document.querySelector(`#${EXTENSION_MARKER_ID}`) !== null) {
        logger.info('Native integration is already running')
        return
    }
    // If the extension marker isn't present, inject it and listen for a custom event sent by the native
    // integration to signal its activation.
    injectExtensionMarker()
    const nativeIntegrationActivationEventReceived = fromEvent(document, NATIVE_INTEGRATION_ACTIVATED)
        .pipe(first())
        .toPromise()

    subscriptions.add(
        await injectCodeIntelligence(getAssetsURL(CLOUD_SOURCEGRAPH_URL), true, async function onCodeHostFound() {
            const styleSheets = [
                {
                    id: 'ext-style-sheet',
                    path: 'css/style.bundle.css',
                },
                {
                    id: 'ext-style-sheet-css-modules',
                    path: 'css/inject.bundle.css',
                },
            ]

            await Promise.all(styleSheets.map(loadStyleSheet).map(waitForStyleSheet))
        })
    )

    // Clean up susbscription if the native integration gets activated
    // later in the lifetime of the content script.
    await nativeIntegrationActivationEventReceived
    logger.info('Native integration activation event received')
    subscriptions.unsubscribe()
}

main().catch(console.error)
