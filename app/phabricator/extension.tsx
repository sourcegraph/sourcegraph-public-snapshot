import '../util/polyfill'

import { setSourcegraphUrl } from '../util/context'
import { getPhabricatorCSS, getSourcegraphURLFromConduit } from './backend'
import { injectPhabricatorBlobAnnotators } from './inject'
import { expanderListen, metaClickOverride, setupPageLoadListener } from './util'

// NOTE: injectModules is idempotent, so safe to call multiple times on the same page.
function injectModules(): void {
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)

    injectPhabricatorBlobAnnotators().catch(e => console.error(e))
}

export function init(): void {
    /**
     * This is the main entry point for the phabricator in-page JavaScript plugin.
     */
    if (window.localStorage && window.localStorage.SOURCEGRAPH_DISABLED !== 'true') {
        document.addEventListener('phabPageLoaded', () => {
            // Backwards compat: Support Legacy Phabricator extension. Check that the Phabricator integration
            // passed the bundle url. Legacy Phabricator extensions inject CSS via the loader.js script
            // so we do not need to do this here.
            if (!window.SOURCEGRAPH_BUNDLE_URL && !window.localStorage.SOURCEGRAPH_BUNDLE_URL) {
                injectModules()
                metaClickOverride()
                expanderListen()
                return
            }

            getSourcegraphURLFromConduit()
                .then(sourcegraphUrl => {
                    getPhabricatorCSS()
                        .then(css => {
                            const style = document.createElement('style') as HTMLStyleElement
                            style.setAttribute('type', 'text/css')
                            style.id = 'sourcegraph-styles'
                            style.textContent = css
                            document.getElementsByTagName('head')[0].appendChild(style)
                            window.localStorage.SOURCEGRAPH_URL = sourcegraphUrl
                            window.SOURCEGRAPH_URL = sourcegraphUrl
                            setSourcegraphUrl(sourcegraphUrl)
                            expanderListen()
                            metaClickOverride()
                            injectModules()
                        })
                        .catch(e => {
                            console.error(e)
                        })
                })
                .catch(e => console.error(e))
        })
        setupPageLoadListener()
    } else {
        // tslint:disable-next-line
        console.log(
            `Sourcegraph on Phabricator is disabled because window.localStorage.SOURCEGRAPH_DISABLED is set to ${
                window.localStorage.SOURCEGRAPH_DISABLED
            }.`
        )
    }
}

const url = window.SOURCEGRAPH_URL || window.localStorage.SOURCEGRAPH_URL
if (url) {
    setSourcegraphUrl(url)
}
init()
