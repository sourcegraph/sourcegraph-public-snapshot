import '../../config/polyfill'

import { setSourcegraphUrl } from '../../shared/util/context'
import { featureFlags } from '../../shared/util/featureFlags'
import { injectCodeIntelligence } from '../code_intelligence/inject'
import { getPhabricatorCSS, getSourcegraphURLFromConduit } from './backend'
import { phabCodeViews } from './code_views'
import { injectPhabricatorBlobAnnotators } from './inject_old'
import { expanderListen, metaClickOverride, setupPageLoadListener } from './util'

// NOTE: injectModules is idempotent, so safe to call multiple times on the same page.
function injectModules(): void {
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)

    featureFlags
        .isEnabled('newTooltips')
        .then(enabled => {
            if (enabled) {
                injectCodeIntelligence({ name: 'phabricator', codeViews: phabCodeViews })
                return
            }

            injectPhabricatorBlobAnnotators().catch(e => console.error(e))
        })
        .catch(err => console.error('could not get feature flag', err))
}

export function init(): void {
    /**
     * This is the main entry point for the phabricator in-page JavaScript plugin.
     */
    if (window.localStorage && window.localStorage.getItem('SOURCEGRAPH_DISABLED') !== 'true') {
        document.addEventListener('phabPageLoaded', () => {
            // Backwards compat: Support Legacy Phabricator extension. Check that the Phabricator integration
            // passed the bundle url. Legacy Phabricator extensions inject CSS via the loader.js script
            // so we do not need to do this here.
            if (!window.SOURCEGRAPH_BUNDLE_URL && !window.localStorage.getItem('SOURCEGRAPH_BUNDLE_URL')) {
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
                            window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphUrl)
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
