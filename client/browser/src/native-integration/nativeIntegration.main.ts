// Set globals first before any imports.
import '../config/page.entry'
// Polyfill before other imports.
import '@sourcegraph/shared/src/polyfills'

import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { injectCodeIntelligence } from '../shared/code-hosts/shared/inject'
import {
    EXTENSION_MARKER_ID,
    injectExtensionMarker,
    NATIVE_INTEGRATION_ACTIVATED,
} from '../shared/code-hosts/sourcegraph/inject'
import { getAssetsURL } from '../shared/util/context'

const IS_EXTENSION = false

setLinkComponent(AnchorLink)

interface InsertStyleSheetOptions {
    id: string
    path: string
    assetsURL: string
}

function insertStyleSheet({ id, path, assetsURL }: InsertStyleSheetOptions): void {
    const link = document.createElement('link')
    link.setAttribute('rel', 'stylesheet')
    link.setAttribute('type', 'text/css')
    link.setAttribute('href', new URL(path, assetsURL).href)
    link.id = id
    document.head.append(link)
}

function init(): void {
    console.log('Sourcegraph native integration is running')
    const sourcegraphURL = window.SOURCEGRAPH_URL
    if (!sourcegraphURL) {
        throw new Error('window.SOURCEGRAPH_URL is undefined')
    }

    const assetsURL = getAssetsURL(sourcegraphURL)

    if (document.querySelector(`#${EXTENSION_MARKER_ID}`) !== null) {
        // If the extension marker already exists, it means the browser extension is currently executing.
        // Dispatch a custom event to signal that browser extension resources should be cleaned up.
        document.dispatchEvent(new CustomEvent<{}>(NATIVE_INTEGRATION_ACTIVATED))
    } else {
        injectExtensionMarker()
    }
    insertStyleSheet({ id: 'sourcegraph-styles', path: 'css/app.bundle.css', assetsURL })
    insertStyleSheet({ id: 'sourcegraph-styles-css-modules', path: 'css/contentPage.main.bundle.css', assetsURL })
    window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphURL)
    window.SOURCEGRAPH_URL = sourcegraphURL
    // TODO handle subscription
    injectCodeIntelligence({ sourcegraphURL, assetsURL }, IS_EXTENSION).catch(error => {
        console.error('Error injecting Sourcegraph code navigation:', error)
    })
}

init()
