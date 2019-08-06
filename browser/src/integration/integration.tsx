import '../config/polyfill'

import * as H from 'history'
import React from 'react'
import { setLinkComponent } from '../../../shared/src/components/Link'
import { injectCodeIntelligence } from '../libs/code_intelligence/inject'
import { EXTENSION_MARKER_ID, injectExtensionMarker, NATIVE_INTEGRATION_ACTIVATED } from '../libs/sourcegraph/inject'

const IS_EXTENSION = false

setLinkComponent(({ to, children, ...props }) => (
    <a href={to && typeof to !== 'string' ? H.createPath(to) : to} {...props}>
        {children}
    </a>
))

async function init(): Promise<void> {
    console.log('Sourcegraph native integration is running')
    const sourcegraphURL = window.SOURCEGRAPH_URL
    if (!sourcegraphURL) {
        throw new Error('window.SOURCEGRAPH_URL is undefined')
    }
    if (document.getElementById(EXTENSION_MARKER_ID) !== null) {
        // If the extension marker already exists, it means the browser extension is currently executing.
        // Dispatch a custom event to signal that browser extension resources should be cleaned up.
        document.dispatchEvent(new CustomEvent<{}>(NATIVE_INTEGRATION_ACTIVATED))
    } else {
        injectExtensionMarker()
    }
    const link = document.createElement('link')
    link.setAttribute('rel', 'stylesheet')
    link.setAttribute('type', 'text/css')
    link.setAttribute('href', sourcegraphURL + '/.assets/extension/css/style.bundle.css')
    link.id = 'sourcegraph-styles'
    document.getElementsByTagName('head')[0].appendChild(link)
    window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphURL)
    window.SOURCEGRAPH_URL = sourcegraphURL
    // TODO handle subscription
    await injectCodeIntelligence(IS_EXTENSION)
}

init().catch(err => {
    console.error('Error initializing integration', err)
})
