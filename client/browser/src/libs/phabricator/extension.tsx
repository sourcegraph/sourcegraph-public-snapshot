import '../../config/polyfill'

import { setSourcegraphUrl } from '../../shared/util/context'
import { injectCodeIntelligence } from '../code_intelligence'
import { getPhabricatorCSS, getSourcegraphURLFromConduit } from './backend'
import { metaClickOverride } from './util'

// NOTE: injectModules is idempotent, so safe to call multiple times on the same page.
async function injectModules(): Promise<void> {
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)

    await injectCodeIntelligence()
}

function init(): void {
    /**
     * This is the main entry point for the phabricator in-page JavaScript plugin.
     */
    if (window.localStorage && window.localStorage.getItem('SOURCEGRAPH_DISABLED') !== 'true') {
        // Backwards compat: Support Legacy Phabricator extension. Check that the Phabricator integration
        // passed the bundle url. Legacy Phabricator extensions inject CSS via the loader.js script
        // so we do not need to do this here.
        if (!window.SOURCEGRAPH_BUNDLE_URL && !window.localStorage.getItem('SOURCEGRAPH_BUNDLE_URL')) {
            injectModules().catch(err => console.error('Unable to inject modules', err))
            metaClickOverride()
            return
        }

        getSourcegraphURLFromConduit()
            .then(sourcegraphUrl => {
                getPhabricatorCSS()
                    .then(css => {
                        const style = document.createElement('style')
                        style.setAttribute('type', 'text/css')
                        style.id = 'sourcegraph-styles'
                        style.textContent = css
                        document.getElementsByTagName('head')[0].appendChild(style)
                        window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphUrl)
                        window.SOURCEGRAPH_URL = sourcegraphUrl
                        setSourcegraphUrl(sourcegraphUrl)
                        metaClickOverride()
                        injectModules().catch(err => console.error('Unable to inject modules', err))
                    })
                    .catch(e => {
                        console.error(e)
                    })
            })
            .catch(e => console.error(e))
    } else {
        // tslint:disable-next-line
        console.log(
            `Sourcegraph on Phabricator is disabled because window.localStorage.getItem('SOURCEGRAPH_DISABLED') is set to ${window.localStorage.getItem(
                'SOURCEGRAPH_DISABLED'
            )}.`
        )
    }
}

const url = window.SOURCEGRAPH_URL || window.localStorage.getItem('SOURCEGRAPH_URL')
if (url) {
    setSourcegraphUrl(url)
}
init()
