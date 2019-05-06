import '../../config/polyfill'

import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { setLinkComponent } from '../../../../shared/src/components/Link'
import { setSourcegraphUrl } from '../../shared/util/context'
import { MutationRecordLike, observeMutations } from '../../shared/util/dom'
import { determineCodeHost, injectCodeIntelligenceToCodeHost } from '../code_intelligence'
import { getPhabricatorCSS, getSourcegraphURLFromConduit } from './backend'
import { metaClickOverride } from './util'

// Just for informational purposes (see getPlatformContext())
window.SOURCEGRAPH_PHABRICATOR_EXTENSION = true

// NOT idempotent.
async function injectModules(): Promise<void> {
    // This is added so that the browser extension doesn't
    // interfere with the native Phabricator integration.
    // TODO this is racy because the script is loaded async
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)

    const mutations: Observable<MutationRecordLike[]> = observeMutations(document.body, {
        childList: true,
        subtree: true,
    }).pipe(startWith([{ addedNodes: [document.body], removedNodes: [] }]))

    // TODO handle subscription
    const codeHost = await determineCodeHost()
    if (codeHost) {
        await injectCodeIntelligenceToCodeHost(mutations, codeHost)
    }
}

setLinkComponent(({ to, children, ...props }) => (
    <a href={to && typeof to !== 'string' ? H.createPath(to) : to} {...props}>
        {children}
    </a>
))

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
